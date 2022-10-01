package ttlock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	ttlockapi "github.com/nikolai5slo/ttlock2mqtt/ttlock-api"
)

func (s *TTLockAPIService) GetLocks(cred Credentials) ([]Lock, error) {
	response, err := s.autoAuth(&cred, func(clientID string, accessToken string) (interface{}, error) {
		listLocksParams := &ttlockapi.ListLocksParams{
			ClientId:    clientID,
			AccessToken: accessToken,
			LockAlias:   nil,
			GroupId:     nil,
			PageNo:      1,
			PageSize:    100,
			Date:        time.Now().UnixMilli(),
		}

		return s.ttlockClient.ListLocksWithResponse(context.TODO(), listLocksParams)
	}, func(i interface{}) []byte { return i.(*ttlockapi.ListLocksResponse).Body })

	if err != nil {
		return nil, err
	}

	var lockList []Lock

	err = mapstructure.Decode((*response.(*ttlockapi.ListLocksResponse).JSON200).(map[string]interface{})["list"], &lockList)

	if err != nil {
		return nil, err
	}

	return lockList, nil
}

func (s *TTLockAPIService) GetLockStatus(cred Credentials, l Lock) (LockStatus, error) {
	response, err := s.autoAuth(&cred, func(clientID string, accessToken string) (interface{}, error) {
		getLockOpenStateParams := &ttlockapi.GetLockOpenStateParams{
			ClientId:    clientID,
			AccessToken: accessToken,
			LockId:      l.LockId,
			Date:        time.Now().UnixMilli(),
		}

		return s.ttlockClient.GetLockOpenStateWithResponse(context.TODO(), getLockOpenStateParams)
	}, func(i interface{}) []byte { return i.(*ttlockapi.GetLockOpenStateResponse).Body })

	if err != nil {
		return Unknown, err
	}

	data := &ttlockapi.LockOpenState{}

	err = json.Unmarshal(response.(*ttlockapi.GetLockOpenStateResponse).Body, &data)

	if err != nil {
		return Unknown, err
	}

	if data.State == nil {
		return Unknown, fmt.Errorf("missing state in the response")
	}

	return LockStatus(*data.State), nil
}

func (s *TTLockAPIService) Lock(cred Credentials, l Lock) error {
	_, err := s.autoAuth(&cred, func(clientID string, accessToken string) (interface{}, error) {
		data := url.Values{}
		data.Add("clientId", s.clientID)
		data.Add("accessToken", cred.AccessToken)
		data.Add("lockId", fmt.Sprint(l.LockId))
		data.Add("date", fmt.Sprint(time.Now().UnixMilli()))

		return s.ttlockClient.PostLockWithBodyWithResponse(context.TODO(), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	}, func(i interface{}) []byte { return i.(*ttlockapi.PostLockResponse).Body })

	return err
}

func (s *TTLockAPIService) Unlock(cred Credentials, l Lock) error {
	_, err := s.autoAuth(&cred, func(clientID string, accessToken string) (interface{}, error) {
		data := url.Values{}
		data.Add("clientId", s.clientID)
		data.Add("accessToken", cred.AccessToken)
		data.Add("lockId", fmt.Sprint(l.LockId))
		data.Add("date", fmt.Sprint(time.Now().UnixMilli()))

		return s.ttlockClient.PostUnlockWithBodyWithResponse(context.TODO(), "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	}, func(i interface{}) []byte { return i.(*ttlockapi.PostUnlockResponse).Body })

	return err
}
