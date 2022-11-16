package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/nikolai5slo/ttlock2mqtt/credentials"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
	"github.com/nikolai5slo/ttlock2mqtt/mqtt"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
)

type Controller struct {
	lockStorage   locks.Storage
	credStorage   credentials.Storage
	mqtt          *mqtt.HAMqtt
	refreshRate   time.Duration
	ttlockService ttlock.Service

	lastRefresh     time.Time
	introducedLocks locks.LockList
}

type Conf func(*Controller) error

func New(cfg ...Conf) (*Controller, error) {
	s := &Controller{
		refreshRate: 60 * time.Second,
	}

	for _, c := range cfg {
		if err := c(s); err != nil {
			return s, fmt.Errorf("configuration failed: %w", err)
		}
	}

	return s, nil
}

func WithLockStorage(l locks.Storage) Conf {
	return func(c *Controller) error {
		c.lockStorage = l
		return nil
	}
}

func WithCredentialsStorage(s credentials.Storage) Conf {
	return func(c *Controller) error {
		c.credStorage = s
		return nil
	}
}

func WithMqtt(m *mqtt.HAMqtt) Conf {
	return func(c *Controller) error {
		c.mqtt = m
		return nil
	}
}

func WithRefreshRate(d time.Duration) Conf {
	return func(c *Controller) error {
		c.refreshRate = d
		return nil
	}
}

func WithTTlockService(t ttlock.Service) Conf {
	return func(c *Controller) error {
		c.ttlockService = t
		return nil
	}
}

func (c *Controller) StartAutoRefresh() {
	go c.runRefresh()
}

func (c *Controller) runRefresh() {
	for {
		c.lastRefresh = time.Now()
		err := c.Refresh()

		if err != nil {
			log.Printf("auto refresh failed: %s", err)
		}

		d := time.Until(c.lastRefresh.Add(c.refreshRate))
		if d > 0 {
			time.Sleep(d)
		}
	}
}

func (c *Controller) Refresh() error {
	// Read locks
	mLocks := locks.LockList{}

	err := c.lockStorage.Load(&mLocks)

	if err != nil {
		return fmt.Errorf("cannot load locks: %w", err)
	}

	// Read credentials
	creds := credentials.CredentialsList{}

	err = c.credStorage.Load(&creds)

	if err != nil {
		return fmt.Errorf("cannot load credentials: %w", err)
	}

	newLocks := mLocks.Diff(c.introducedLocks)

	// Introduce new locks
	for _, l := range newLocks {
		err := c.mqtt.IntroduceLock(l)
		if err != nil {
			return fmt.Errorf("was not able to update the %d lock on mqtt: %w", l.LockId, err)
		}

		getLockCallback := func(lck locks.ManagedLock) func(ls ttlock.LockStatus) {
			return func(ls ttlock.LockStatus) {
				cred := creds.Get(lck.CredentialsID)

				if cred == nil {
					log.Printf("cannot find credentials: %d", lck.CredentialsID)
				}

				var err error

				if ls == ttlock.Locked {
					err = c.ttlockService.Lock(*cred, lck.Lock)
				} else if ls == ttlock.Unlocked {
					err = c.ttlockService.Unlock(*cred, lck.Lock)
				}

				if err != nil {
					log.Printf("failed to lock/unlock [%d]: %s", lck.LockId, err)
				} else {
					err = c.mqtt.UpdateLockStatus(lck, ls)

					if err != nil {
						log.Printf("failed to update lock status: %s", err)
					}
				}
			}
		}

		// Monitor lock commands
		if err := c.mqtt.MqttLockCommandCallback(l, getLockCallback(l)); err != nil {
			log.Printf("Failed to monitor lock: %s", err)
		} else {
			log.Printf("introduced new lock: %d", l.LockId)
			c.introducedLocks = c.introducedLocks.Add(l)
		}
	}

	///
	// Get lock statuses
	//
	for i, l := range c.introducedLocks {
		cred := creds.Get(l.CredentialsID)

		if cred == nil {
			return fmt.Errorf("cannot find credentials: %d", l.CredentialsID)
		}

		status, err := c.ttlockService.GetLockStatus(*cred, l.Lock)

		if err != nil {
			log.Printf("cannot get lock status [%d]: %s", l.LockId, err)
		}

		err = c.mqtt.UpdateLockStatus(l, status)

		if err != nil {
			log.Printf("failed to update lock status: %s", err)
		}

		if i < len(c.introducedLocks)-1 {
			time.Sleep(time.Duration(int(c.refreshRate) / len(c.introducedLocks)))
		}
	}

	return nil
}

func (c *Controller) Close() error {
	c.mqtt.Close()
	return nil
}
