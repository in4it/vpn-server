package oidcrenewal

import (
	"fmt"
	"log"
	"time"

	"github.com/in4it/wireguard-server/pkg/logging"
)

const WAKEUP_TIME_SECONDS = 300         // every 5 minutes we check
const DEFAULT_RENEWAL_TIME_MINUTES = 60 // every hour we want to refresh the token
const RENEWAL_RETRIES = 3               // 3 retries before we suspend a user
const RENEWAL_BACKOFF_SECONDS = 10      // time between requests to the oidc providers

func (r *Renewal) Worker() {
	// do renewal
	if r.enabled {
		fmt.Printf("Starting oidc renewal worker (loglevel: %d)\n", logging.Loglevel)
	}
	for {
		if !r.enabled {
			time.Sleep(WAKEUP_TIME_SECONDS * time.Second)
			continue
		}
		deletedEntries := r.oidcStore.CleanupOAuth2DataForAllEntries()
		if deletedEntries > 0 {
			err := r.oidcStore.SaveOIDCStore()
			if err != nil {
				log.Printf("Renewal Worker: [warning] couldn't save oidc store after cleanup: %s", err)
			}
		}
		for key, oauth2Data := range r.oidcStore.OAuth2Data {
			logging.DebugLog(fmt.Errorf("running canRenew of %s", oauth2Data.ID))
			// can we renew? Do we have expiration date and it is expired?
			canRenew, oidcProvider, discovery, err := canRenew(r.renewalTime, oauth2Data, r.oidcStore, r.oidcProviders)
			if err != nil {
				log.Printf("Renewal Worker: [warning] needsRenewal: %s", err)
			}
			if canRenew {
				logging.DebugLog(fmt.Errorf("we can renew %s", oauth2Data.ID))
				r.renew(discovery, key, oauth2Data, oidcProvider) // error logging within function
			}
			time.Sleep(RENEWAL_BACKOFF_SECONDS * time.Second)
		}
		time.Sleep(WAKEUP_TIME_SECONDS * time.Second)
	}
}
