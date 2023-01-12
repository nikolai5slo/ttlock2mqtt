package ttlock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	ttlockapi "github.com/nikolai5slo/ttlock2mqtt/ttlock-api"
)

func (s *TTLockAPIService) Login(username string, password string) (Credentials, error) {
	cred := Credentials{}

	// Get Refresh Token
	data := url.Values{}
	data.Add("clientId", s.clientID)
	data.Add("clientSecret", s.clientSecret)
	data.Add("username", username)
	data.Add("password", password)

	response, err := s.ttlockClient.GetTokenWithBodyWithResponse(context.TODO(), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))

	if err != nil {
		return cred, err
	}

	var errorResponse ttlockapi.Error

	err = json.Unmarshal(response.Body, &errorResponse)
	if err == nil && errorResponse.Errmsg != "" {
		return cred, fmt.Errorf(errorResponse.Errmsg)
	}

	var credentialsResponse ttlockapi.Credentials
	err = json.Unmarshal(response.Body, &credentialsResponse)
	if err != nil {
		return cred, err
	}

	if credentialsResponse.AccessToken == "" || credentialsResponse.RefreshToken == "" {
		return cred, fmt.Errorf("Credentials not present in response")
	}

	cred.AccessToken = credentialsResponse.AccessToken
	cred.RefreshToken = credentialsResponse.RefreshToken
	cred.ID = credentialsResponse.Uid
	cred.Username = username
	cred.ExpiresAt = time.Now().Add(time.Duration(credentialsResponse.ExpiresIn) * time.Second)

	return cred, nil
}

func (s *TTLockAPIService) refreshToken(cred *Credentials) error {
	// Get Refresh Token
	data := url.Values{}
	data.Add("clientId", s.clientID)
	data.Add("clientSecret", s.clientSecret)
	data.Add("grant_type", "refresh_token")
	data.Add("refresh_token", cred.RefreshToken)

	response, err := s.ttlockClient.GetTokenWithBodyWithResponse(context.TODO(), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))

	if err != nil {
		return err
	}

	var errorResponse ttlockapi.Error

	err = json.Unmarshal(response.Body, &errorResponse)
	if err == nil && errorResponse.Errmsg != "" {
		return fmt.Errorf(errorResponse.Errmsg)
	}

	var credentialsResponse ttlockapi.Credentials
	err = json.Unmarshal(response.Body, &credentialsResponse)
	if err != nil {
		return err
	}

	if credentialsResponse.AccessToken != "" || credentialsResponse.RefreshToken != "" {
		return fmt.Errorf("credentials not present in response")
	}

	if cred.ID != credentialsResponse.Uid {
		return fmt.Errorf("credentials refresh mismatch")

	}

	cred.AccessToken = credentialsResponse.AccessToken
	cred.RefreshToken = credentialsResponse.RefreshToken
	cred.ExpiresAt = time.Now().Add(time.Duration(credentialsResponse.ExpiresIn) * time.Second)

	return nil
}

func (s *TTLockAPIService) autoAuth(cred *Credentials, fn func(string, string) (interface{}, error), getBody func(interface{}) []byte, retryCount int) (interface{}, error) {
	response, err := fn(s.clientID, cred.AccessToken)

	if err != nil {
		return response, err
	}

	var errorResponse ttlockapi.Error

	body := getBody(response)
	err = json.Unmarshal(body, &errorResponse)

	// If no error code
	if err != nil || errorResponse.Errcode == nil || *errorResponse.Errcode == 0 {
		return response, err
	}

	// If token refresh error
	if *errorResponse.Errcode == 10003 {
		// Do token refresh
		err = s.refreshToken(cred)
		if err != nil {
			return response, fmt.Errorf("token refresh failed: %w", err)
		}

		// Retry request
		return fn(s.clientID, cred.AccessToken)
	}

	// Failed
	if *errorResponse.Errcode == 1 {
		for retryCount > 0 {
			log.Print("Retrying")
			response, err = fn(s.clientID, cred.AccessToken)

			if err == nil {
				break
			}
			retryCount--
		}
	}

	return response, fmt.Errorf("error response: \"%s\" Body: %s", errorResponse.Errmsg, string(body))
}
