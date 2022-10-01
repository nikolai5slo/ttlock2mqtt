package locks

import (
	"os"

	"github.com/schollz/jsonstore"
)

type JsonStore struct {
	store    *jsonstore.JSONStore
	filePath string
}

func NewJsonStore(store *jsonstore.JSONStore, filePath string) (*JsonStore, error) {
	return &JsonStore{
		store:    store,
		filePath: filePath,
	}, nil
}

func (s *JsonStore) Save(l LockList) error {
	err := s.store.Set("locks", l)

	if err != nil {
		return err
	}

	return jsonstore.Save(s.store, s.filePath)
}

func (s *JsonStore) Load(l *LockList) error {
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return nil
	}

	js, err := jsonstore.Open(s.filePath)

	if err != nil {
		return err
	}

	s.store = js

	err = js.Get("locks", l)

	// Ignore no souch key errors
	if _, ok := err.(jsonstore.NoSuchKeyError); ok {
		return nil
	}

	return err
}
