package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/juli3nk/go-freebox"
)

func getAppToken(dev *freebox.Device, appID string) (*string, error) {
	var appToken string

	appTokenBytes, err := getAppTokenFromFile()
	if err != nil {
		tokenReq := freebox.TokenRequest{
			AppID:      appID,
			AppName:    "Presence Tracker",
			AppVersion: "0.1.0",
			DeviceName: "Kaza",
		}

		jsonData, err := json.Marshal(tokenReq)
		if err != nil {
			return nil, err
		}

		response, err := dev.RequestAuthorization(bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, err
		}

		c := time.Tick(10 * time.Second)
		for _ = range c {
			status, err := dev.TrackAuthorizationProgress(response.Result.TrackID)
			if err != nil {
				return nil, err
			}

			if status.Result.Status == freebox.AuthorizationStatusGranted {
				break
			}
		}

		if err := saveAppTokenToFile(response.Result.AppToken); err != nil {
			return nil, err
		}

		appTokenBytes, err = getAppTokenFromFile()
		if err != nil {
			return nil, err
		}
	}
	appToken = string(appTokenBytes)

	return &appToken, nil
}

func getSessionToken(dev *freebox.Device, appID, appToken string) (*string, error) {
	var sessionToken string

	sessionTokenBytes, err := getSessionTokenFromFile()
	if err != nil {
		challenge, err := dev.GetChallenge()
		if err != nil {
			return nil, err
		}

		session, err := dev.OpenSession(appID, appToken, challenge.Result.Challenge)
		if err != nil {
			return nil, err
		}

		if err := saveSessionTokenToFile(session.Result.SessionToken); err != nil {
			return nil, err
		}

		sessionTokenBytes, err = getSessionTokenFromFile()
		if err != nil {
			return nil, err
		}
	}
	sessionToken = strings.TrimSpace(string(sessionTokenBytes))

	return &sessionToken, nil
}
