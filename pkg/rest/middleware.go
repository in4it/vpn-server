package rest

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/in4it/wireguard-server/pkg/auth/oidc"
	"github.com/in4it/wireguard-server/pkg/users"
)

type CustomValue string

// auth middleware

func (c *Context) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.SetupCompleted {
			c.returnError(w, fmt.Errorf("setup not completed"), http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			c.writeWithStatus(w, []byte(`{"error": "token not found"}`), http.StatusUnauthorized)
			return
		}
		tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", -1)
		if len(tokenString) == 0 {
			c.returnError(w, fmt.Errorf("empty token"), http.StatusUnauthorized)
			return
		}

		// determine token to parse
		var tokenToParse string
		// is token an access token or a jwt from local auth?
		kid, _ := getKidFromToken(tokenString)
		if kid == c.JWTKeysKID { // local auth token
			tokenToParse = tokenString
		} else {
			for _, oauth2Data := range c.OIDCStore.OAuth2Data {
				if oauth2Data.Token.AccessToken == tokenString {
					tokenToParse = oauth2Data.Token.IDToken
				}
			}
			if tokenToParse == "" {
				c.returnError(w, fmt.Errorf("token error: access token not found (wrong token or token expired)"), http.StatusUnauthorized)
				return
			}
		}
		token, err := jwt.Parse(tokenToParse, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Header["kid"]; !ok {
				return nil, fmt.Errorf("no kid header found in token")
			}
			if token.Header["kid"] == c.JWTKeysKID {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("local kid: unexpected signing method: %v", token.Header["alg"])
				}
				return c.JWTKeys.PublicKey, nil
			}
			discoveryProviders := make([]oidc.Discovery, len(c.OIDCProviders))
			for k, oidcProvider := range c.OIDCProviders {
				discovery, err := c.OIDCStore.GetDiscoveryURI(oidcProvider.DiscoveryURI)
				if err != nil {
					return nil, fmt.Errorf("couldn't retrieve discoveryURI from OIDC Provider (check discovery URI in OIDC settings). Error: %s", err)
				}
				discoveryProviders[k] = discovery
			}
			allJwks, err := c.OIDCStore.GetAllJwks(discoveryProviders)
			if err != nil {
				return nil, fmt.Errorf("couldn't retrieve JWKS URL from OIDC Provider (check discovery URI in OIDC settings). Error: %s", err)
			}
			publicKey, err := oidc.GetPublicKeyForToken(allJwks, discoveryProviders, token)
			if err != nil {
				return nil, fmt.Errorf("GetPublicKeyForToken error: %s", err)
			}
			return publicKey, nil
		})
		if err != nil {
			c.returnError(w, fmt.Errorf("token error: %s", err), http.StatusUnauthorized)
			return
		}
		token.Claims.(jwt.MapClaims)["kid"] = token.Header["kid"]
		ctx := context.WithValue(r.Context(), CustomValue("claims"), token.Claims.(jwt.MapClaims))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// logging middleware

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
// MIT licensed
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (c *Context) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedResponse := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(wrappedResponse, r)
		log.Printf("req=%s res=%d method=%s src=%s duration=%s", r.RequestURI, wrappedResponse.status, r.Method, r.RemoteAddr, time.Since(start))
	})
}

func (c *Context) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			sendCorsHeaders(w, r.Header.Get("Access-Control-Request-Headers"), c.Hostname, c.Protocol)
			w.WriteHeader(http.StatusNoContent)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
func (c *Context) injectUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := c.GetUserFromRequest(r)
		if err != nil {
			c.returnError(w, fmt.Errorf("token error: %s", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CustomValue("user"), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *Context) isAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(CustomValue("user")).(users.User)
		if user.Role != "admin" {
			c.returnError(w, fmt.Errorf("endpoint forbidden"), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (c *Context) httpsRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.RedirectToHttps && r.TLS == nil {
			if strings.HasPrefix(r.URL.Path, "/api") {
				c.returnError(w, fmt.Errorf("non-tls requests disabled"), http.StatusForbidden)
				return
			}
			http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, r.RequestURI), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (c *Context) isSCIMEnabled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.SCIM.EnableSCIM {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{ "error": "SCIM Not Enabled" }`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
