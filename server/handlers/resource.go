package handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nikolai5slo/ttlock2mqtt/credentials"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
)

type Resource struct {
	h *Handlers
	c *gin.Context
}

func (h *Handlers) Res(c *gin.Context) *Resource {
	return &Resource{
		h: h,
		c: c,
	}
}

func (r *Resource) GetCredentials() (creds credentials.CredentialsList, err error) {
	err = r.h.credStorage.Load(&creds)
	return
}

func (r *Resource) GetManagedLocks() (locks locks.LockList, err error) {
	err = r.h.lockStorage.Load(&locks)
	return
}

func (r *Resource) GetSelectedCredentials() (cred *credentials.Credentials, err error) {
	credID, err := strconv.Atoi(r.c.PostForm("credentials"))

	if err != nil {
		return
	}

	creds, err := r.GetCredentials()

	if err != nil {
		return
	}

	cred = creds.Get(int32(credID))
	if cred == nil {
		return nil, fmt.Errorf("cannot find credentials for the ID: %d", credID)
	}

	return cred, nil
}

func (r *Resource) GetSelectedLocks() (mLocks []ttlock.Lock, err error) {

	cred, err := r.GetSelectedCredentials()

	if err != nil {
		return
	}

	l, err := r.h.ttlockService.GetLocks(*cred)

	if err != nil {
		return
	}

	err = r.c.Request.ParseForm()

	if err != nil {
		return
	}

	// Loop through checkboxes
	var lockIDs []int32
	for _, k := range r.c.PostFormArray("locks") {
		lockID, err := strconv.Atoi(k)
		if err != nil {
			log.Printf("failed to parse lock ID")
			continue
		}
		lockIDs = append(lockIDs, int32(lockID))
	}

	// Get selected locks
	for _, lockID := range lockIDs {
		for _, lock := range l {
			if lock.LockId == lockID {
				mLocks = append(mLocks, lock)
			}
		}
	}

	return
}
