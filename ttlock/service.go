package ttlock

type Service interface {
	GetLocks(Credentials) ([]Lock, error)
	Login(string, string) (Credentials, error)
	GetLockStatus(cred Credentials, l Lock) (LockStatus, error)
	Lock(cred Credentials, l Lock) error
	Unlock(cred Credentials, l Lock) error
}
