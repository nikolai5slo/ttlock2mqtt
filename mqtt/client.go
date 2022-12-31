package mqtt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nikolai5slo/ttlock2mqtt/locks"
	"github.com/nikolai5slo/ttlock2mqtt/ttlock"
)

type HAMqtt struct {
	opts    *mqtt.ClientOptions
	client  mqtt.Client
	timeout time.Duration
}

type MqttLockConfig struct {
	CommandTopic string     `json:"command_topic"`
	StateTopic   string     `json:"state_topic"`
	Name         string     `json:"name"`
	UniqueID     string     `json:"unique_id"`
	Device       MqttDevice `json:"device"`
}
type MqttDevice struct {
	Name        string   `json:"name"`
	Model       string   `json:"model"`
	Identifiers []string `json:"identifiers"`
}

type Conf func(*HAMqtt) error

func New(cfg ...Conf) (*HAMqtt, error) {
	mqt := &HAMqtt{
		timeout: 2 * time.Second,
	}

	mqt.opts = mqtt.NewClientOptions()
	mqt.opts.SetAutoReconnect(true)

	for _, c := range cfg {
		if err := c(mqt); err != nil {
			return mqt, fmt.Errorf("mqtt configuration failed: %w", err)
		}
	}

	mqt.client = mqtt.NewClient(mqt.opts)

	return mqt, nil
}

func WithBroker(broker string) Conf {
	return func(h *HAMqtt) error {
		h.opts.AddBroker(broker)
		return nil
	}
}

func WithClientID(clientID string) Conf {
	return func(h *HAMqtt) error {
		h.opts.SetClientID(clientID)
		return nil
	}
}

func WithCredentials(username string, password string) Conf {
	return func(h *HAMqtt) error {
		h.opts.SetUsername(username)
		h.opts.SetPassword(password)
		return nil
	}
}

func (m *HAMqtt) Connect() error {
	if !m.client.IsConnected() {
		token := m.client.Connect()
		if !token.WaitTimeout(m.timeout) {
			return fmt.Errorf("mqtt connection timeout: %w", token.Error())
		}
		return token.Error()
	}
	return nil
}

func (m *HAMqtt) Close() error {
	m.client.Disconnect(1)
	return nil
}

func (m *HAMqtt) MqttLockCommandCallback(l locks.ManagedLock, callback func(ttlock.LockStatus)) error {
	token := m.client.Subscribe(fmt.Sprintf("ttlock2mqtt/%d/command", l.LockId), 1, func(c mqtt.Client, m mqtt.Message) {
		switch string(m.Payload()) {
		case "LOCK":
			callback(ttlock.Locked)
		case "UNLOCK":
			callback(ttlock.Unlocked)
		}
	})

	token.WaitTimeout(m.timeout)

	return token.Error()
}

// Introduce
func (m *HAMqtt) IntroduceLock(l locks.ManagedLock) error {
	lockConfig := &MqttLockConfig{
		CommandTopic: fmt.Sprintf("ttlock2mqtt/%d/command", l.LockId),
		StateTopic:   fmt.Sprintf("ttlock2mqtt/%d/state", l.LockId),
		Name:         l.LockAlias,
		UniqueID:     fmt.Sprint(l.LockId),
		Device: MqttDevice{
			Name:        l.LockAlias,
			Model:       l.LockName,
			Identifiers: []string{*l.LockMac, fmt.Sprint(l.LockId)},
		},
	}

	payload, err := json.Marshal(lockConfig)

	if err != nil {
		return fmt.Errorf("could not serialize lock config object: %w", err)
	}

	return m.handleError(3, func() error {
		token := m.client.Publish(fmt.Sprintf("homeassistant/lock/ttlock2mqtt/%d/config", l.LockId), 1, true, string(payload))

		token.WaitTimeout(1 * m.timeout)

		return token.Error()
	})
}

func (m *HAMqtt) UpdateLockStatus(l locks.ManagedLock, status ttlock.LockStatus) error {
	txtStatus := ""

	switch status {
	case ttlock.Locked:
		txtStatus = "LOCKED"
	case ttlock.Unlocked:
		txtStatus = "UNLOCKED"
	}

	if txtStatus != "" {
		return m.handleError(3, func() error {
			token := m.client.Publish(fmt.Sprintf("ttlock2mqtt/%d/state", l.LockId), 1, false, txtStatus)

			token.WaitTimeout(1 * m.timeout)

			return token.Error()
		})
	}

	return nil
}

func (m *HAMqtt) handleError(retryCount int, closure func() error) error {
	err := errors.New("no execution")

	for retryCount >= 0 && err != nil {
		err = closure()

		if err != nil && strings.Contains(strings.ToLower(err.Error()), "not connected") {
			err = m.Connect()

			// Do not retry if reconnection failed
			if err != nil {
				return err
			}
		}

		retryCount--
	}

	return err
}
