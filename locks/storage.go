package locks

import "github.com/nikolai5slo/ttlock2mqtt/ttlock"

type ManagedLock struct {
	ttlock.Lock
	CredentialsID int32
}

type LockList []ManagedLock

func (l LockList) Find(ID int32) int {
	// Check if exists
	for i, c := range l {
		if c.LockId == ID {
			return i
		}
	}

	return -1
}

func (l LockList) Add(locks ...ManagedLock) LockList {
	nl := make(LockList, len(l))
	copy(nl, l)
	for _, c := range locks {
		i := nl.Find(c.LockId)

		if i > -1 {
			nl[i] = c
		} else {
			nl = append(nl, c)
		}
	}
	return nl
}

// Diff get locks that are present in list and are not in list1
func (list LockList) Diff(list1 LockList) LockList {
	newList := LockList{}

	for _, l := range list {
		if list1.Find(l.LockId) < 0 {
			newList = append(newList, l)
		}
	}
	return newList
}

type Storage interface {
	Save(LockList) error
	Load(*LockList) error
}
