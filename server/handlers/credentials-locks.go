package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikolai5slo/ttlock2mqtt/credentials"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
)

func (h *Handlers) getLocksForCredentials() gin.HandlerFunc {
	return func(c *gin.Context) {
		var errors []string

		sid, _ := c.Params.Get("id")
		lockID, err := strconv.Atoi(sid)
		if err != nil {
			c.Redirect(301, "/credentials")
		}

		// Get all credentials
		creds := credentials.CredentialsList{}

		err = h.credStorage.Load(&creds)
		if err != nil {
			log.Printf("Loading credentials failed: %s", err)
			errors = append(errors, "Internal server error. Check server logs.")

			h.rednerCredentials(c, creds, errors)
			return
		}

		// Find credentials
		cred := creds.Get(int32(lockID))

		if c == nil {
			c.Redirect(301, "/credentials")
		}

		locks, err := h.ttlockService.GetLocks(*cred)

		if err != nil {
			log.Printf("Getting locks from API failed: %s", err)
			errors = append(errors, "Internal server error. Check server logs.")

			h.rednerCredentials(c, creds, errors)
			return
		}

		h.rednerCredentialsWithLocksModal(c, creds, cred, locks, errors)
	}
}

func (h *Handlers) rednerCredentialsWithLocksModal(c *gin.Context, creds credentials.CredentialsList, cred *credentials.Credentials, locks []ttlock.Lock, errors []string) {
	c.HTML(http.StatusOK, "credentials.html", gin.H{
		"credentials": creds,
		"errors":      errors,
		"modal":       true,
		"locks":       locks,
		"credID":      cred.ID,
	})
}
