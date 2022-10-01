package ttlock

import (
	"time"

	ttlockapi "github.com/nikolai5slo/ttlock2mqtt/ttlock-api"
)

type LockStatus int

const (
	Locked   LockStatus = 0
	Unlocked LockStatus = 1
	Unknown  LockStatus = 2
)

type Credentials struct {
	Username     string
	ID           int32
	RefreshToken string
	AccessToken  string
	ExpiresAt    time.Time
}

type Lock = ttlockapi.Lock
