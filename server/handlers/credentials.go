package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikolai5slo/ttlock2mqtt/credentials"
)

func (h *Handlers) registerCredentials(e *gin.Engine) {
	e.POST("/credentials", h.postCredentials())
	e.GET("/credentials", h.getCredentials())
	e.GET("/credentials/:id/locks", h.getLocksForCredentials())
}

func (h *Handlers) getCredentials() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := h.Res(c)

		errors := []string{}

		creds, err := r.GetCredentials()

		if err != nil {
			log.Printf("Loading credentials failed: %s", err)
			h.renderInternalError(c, err)
			return
		}

		h.rednerCredentials(c, creds, errors)
	}
}

func (h *Handlers) postCredentials() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := h.Res(c)

		errors := []string{}

		creds, err := r.GetCredentials()

		if err != nil {
			log.Printf("Loading credentials failed: %s", err)
			h.renderInternalError(c, err)
			return
		}

		username := c.PostForm("username")
		password := c.PostForm("password")

		// Convert password to md5
		hash := md5.Sum([]byte(password))
		password = hex.EncodeToString(hash[:])

		cred, err := h.ttlockService.Login(username, password)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Login failed: %s", err))
			h.rednerCredentials(c, creds, errors)
			return
		}

		newCreds := creds.Add(cred)

		err = h.credStorage.Save(newCreds)
		if err != nil {
			log.Printf("Saving credentials failed: %s", err)
			errors = append(errors, "Saving failed.")

			h.rednerCredentials(c, creds, errors)
			return
		}

		h.rednerCredentials(c, newCreds, errors)
	}
}

func (h *Handlers) rednerCredentials(c *gin.Context, creds credentials.CredentialsList, errors []string) {
	c.HTML(http.StatusOK, "credentials.html", gin.H{
		"credentials": creds,
		"errors":      errors,
		"modal":       false,
	})
}
