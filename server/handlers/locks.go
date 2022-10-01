package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
)

func (h *Handlers) registerLocks(e *gin.Engine) {
	e.GET("/locks", h.getLocks())
	e.POST("/locks", h.postLocks())
}

func (h *Handlers) getLocks() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := h.Res(c)

		errors := []string{}

		l, err := r.GetManagedLocks()

		if err != nil {
			h.renderInternalError(c, err)
			return
		}

		h.renderLocks(c, l, errors)
	}
}

func (h *Handlers) postLocks() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := h.Res(c)

		errors := []string{}

		managedLocks, err := r.GetManagedLocks()

		if err != nil {
			h.renderInternalError(c, err)
			return
		}

		cred, err := r.GetSelectedCredentials()

		if err != nil {
			h.renderInternalError(c, err)
			return
		}

		selectedLocks, err := r.GetSelectedLocks()

		if err != nil {
			h.renderInternalError(c, err)
			return
		}

		for _, selectedLock := range selectedLocks {
			managedLocks = managedLocks.Add(locks.ManagedLock{
				Lock:          selectedLock,
				CredentialsID: cred.ID,
			})
		}

		err = h.lockStorage.Save(managedLocks)

		if err != nil {
			errors = append(errors, "Failed to save locks")
		}

		h.renderLocks(c, managedLocks, errors)
	}
}

func (h *Handlers) renderLocks(c *gin.Context, l locks.LockList, errors []string) {
	c.HTML(http.StatusOK, "locks.html", gin.H{
		"locks":  l,
		"errors": errors,
	})
}
