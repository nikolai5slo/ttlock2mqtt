package ttlock

import ttlockapi "github.com/nikolai5slo/ttlock2mqtt/ttlock-api"

type TTLockAPIService struct {
	ttlockClient ttlockapi.ClientWithResponsesInterface
	clientID     string
	clientSecret string
}

type Conf func(*TTLockAPIService) error

func WithTTLockClient(client ttlockapi.ClientWithResponsesInterface) Conf {
	return func(s *TTLockAPIService) error {
		s.ttlockClient = client
		return nil
	}
}

func WithClientSecret(clientID string, clientSecret string) Conf {
	return func(s *TTLockAPIService) error {
		s.clientID = clientID
		s.clientSecret = clientSecret
		return nil
	}
}

func New(conf ...Conf) (*TTLockAPIService, error) {
	service := &TTLockAPIService{}

	for _, c := range conf {
		if err := c(service); err != nil {
			return service, err
		}
	}

	return service, nil
}
