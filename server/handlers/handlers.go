package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/schollz/jsonstore"
	"github.com/nikolai5slo/ttlock2mqtt/credentials"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
)

type Handlers struct {
	credStorage   credentials.Storage
	lockStorage   locks.Storage
	ttlockService ttlock.Service
}

type Conf func(*Handlers) error

func WithTTlockService(service ttlock.Service) Conf {
	return func(h *Handlers) error {
		h.ttlockService = service
		return nil
	}
}

func WithCredentialsStorage(storage credentials.Storage) Conf {
	return func(h *Handlers) error {
		h.credStorage = storage
		return nil
	}
}

func WithLockStorage(lockStorage locks.Storage) Conf {
	return func(h *Handlers) error {
		h.lockStorage = lockStorage
		return nil
	}
}

func WithStoreFile(filePath string) Conf {
	return func(h *Handlers) error {
		// Credentials store
		cs, err := credentials.NewJsonStore(new(jsonstore.JSONStore), filePath)

		if err != nil {
			return err
		}

		h.credStorage = cs

		// Lock store
		ls, err := locks.NewJsonStore(new(jsonstore.JSONStore), filePath)

		if err != nil {
			return err
		}

		h.lockStorage = ls

		return nil
	}
}

func New(cfg ...Conf) (*Handlers, error) {
	h := &Handlers{}
	for _, c := range cfg {
		if err := c(h); err != nil {
			return h, err
		}
	}

	return h, nil
}

func (h *Handlers) Register(e *gin.Engine) {
	h.registerIndex(e)
	h.registerCredentials(e)
	h.registerLocks(e)
}
