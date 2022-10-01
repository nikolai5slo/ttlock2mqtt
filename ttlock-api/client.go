package ttlockapi

import "net/http"

//go:generate oapi-codegen --config codegen.yaml specs.yaml

type AuthClientWrapper struct {
	client HttpRequestDoer
}

func NewAuthClientWrapper(clientId string, clientSecret string, c Credentials) (*AuthClientWrapper, error) {
	return &AuthClientWrapper{
		client: &http.Client{},
	}, nil
}

func (c *AuthClientWrapper) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
