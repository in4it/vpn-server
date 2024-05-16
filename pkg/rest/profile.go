package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/in4it/wireguard-server/pkg/mfa/totp"
	"github.com/in4it/wireguard-server/pkg/users"
)

func (c *Context) profilePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(CustomValue("user")).(users.User)
	switch r.Method {
	case http.MethodPost:
		var userInput users.User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&userInput)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode input error: %s", err), http.StatusBadRequest)
			return
		}
		if userInput.Password == "" {
			c.returnError(w, fmt.Errorf("no password supplied"), http.StatusBadRequest)
			return
		}
		err = c.UserStore.UpdatePassword(user.ID, userInput.Password)
		if err != nil {
			c.returnError(w, fmt.Errorf("update password error: %s", err), http.StatusBadRequest)
			return
		}

		c.write(w, []byte(`{"result": "OK"}`))
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)

	}
}

func (c *Context) profileFactorsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(CustomValue("user")).(users.User)
	switch r.Method {
	case http.MethodGet:
		factors := make([]users.Factor, len(user.Factors))
		copy(factors, user.Factors)
		for k := range factors {
			factors[k].Secret = "" // remove secret when outputting
		}
		out, err := json.Marshal(factors)
		if err != nil {
			c.returnError(w, fmt.Errorf("factors marshal error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodPost:
		var factor FactorRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&factor)
		if err != nil {
			c.returnError(w, fmt.Errorf("decode factor error: %s", err), http.StatusBadRequest)
			return
		}
		if factor.Secret == "" {
			c.returnError(w, fmt.Errorf("no factor secret supplied"), http.StatusBadRequest)
			return
		}
		if factor.Name == "" {
			c.returnError(w, fmt.Errorf("no factor name supplied"), http.StatusBadRequest)
			return
		}
		if len(factor.Name) > 16 {
			c.returnError(w, fmt.Errorf("factor name too long"), http.StatusBadRequest)
			return
		}
		if factor.Type == "" {
			c.returnError(w, fmt.Errorf("no factor type supplied"), http.StatusBadRequest)
			return
		}
		if factor.Code == "" {
			c.returnError(w, fmt.Errorf("no factor code supplied"), http.StatusBadRequest)
			return
		}

		ok, err := totp.VerifyMultipleIntervals(factor.Secret, factor.Code, 20)
		if err != nil {
			c.returnError(w, fmt.Errorf("totp verify error: %s", err), http.StatusBadRequest)
			return
		}

		if !ok {
			c.returnError(w, fmt.Errorf("code doesn't match. Try entering code again or try with a new QR code"), http.StatusBadRequest)
			return
		}

		user.Factors = append(user.Factors, users.Factor{Type: factor.Type, Secret: factor.Secret, Name: factor.Name})
		out, err := json.Marshal(user.Factors)
		if err != nil {
			c.returnError(w, fmt.Errorf("factors marshal error: %s", err), http.StatusBadRequest)
			return
		}
		err = c.UserStore.UpdateUser(user)
		if err != nil {
			c.returnError(w, fmt.Errorf("coudn't update user: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	case http.MethodDelete:
		factorName := r.PathValue("name")
		if factorName == "" {
			c.returnError(w, fmt.Errorf("no factor name supplied"), http.StatusBadRequest)
			return
		}
		toDelete := -1
		for k := range user.Factors {
			if user.Factors[k].Name == factorName {
				toDelete = k
			}
		}
		if toDelete == -1 {
			c.returnError(w, fmt.Errorf("factor not found"), http.StatusBadRequest)
			return
		}
		user.Factors = append(user.Factors[:toDelete], user.Factors[toDelete+1:]...)
		err := c.UserStore.UpdateUser(user)
		if err != nil {
			c.returnError(w, fmt.Errorf("coudn't update user: %s", err), http.StatusBadRequest)
			return
		}
		out, err := json.Marshal(user.Factors)
		if err != nil {
			c.returnError(w, fmt.Errorf("factors marshal error: %s", err), http.StatusBadRequest)
			return
		}
		c.write(w, out)
	default:
		c.returnError(w, fmt.Errorf("method not supported"), http.StatusBadRequest)

	}
}
