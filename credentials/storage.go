package credentials

import "github.com/nikolai5slo/ttlock2mqtt/ttlock"

type Credentials = ttlock.Credentials

type CredentialsList []Credentials

func (l CredentialsList) Find(ID int32) int {
	// Check if exists
	for i, c := range l {
		if c.ID == ID {
			return i
		}
	}

	return -1
}

func (l CredentialsList) Get(ID int32) *Credentials {
	idx := l.Find(ID)
	if idx >= 0 {
		return &l[idx]
	}
	return nil
}

func (l CredentialsList) Add(creds ...Credentials) CredentialsList {
	nl := make(CredentialsList, len(l))
	copy(nl, l)
	for _, c := range creds {
		i := nl.Find(c.ID)

		if i > -1 {
			nl[i] = c
		} else {
			nl = append(nl, c)
		}
	}
	return nl
}

type Storage interface {
	Save(CredentialsList) error
	Load(*CredentialsList) error
}
