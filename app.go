package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/juli3nk/go-freebox"
)

func getAppToken(st *state, dev *freebox.Device, appID string) error {
	if _, err := st.getAppTokenFromFile(); err != nil {
		tokenReq := freebox.TokenRequest{
			AppID:      appID,
			AppName:    "Presence Tracker",
			AppVersion: "0.1.0",
			DeviceName: "Kaza",
		}

		jsonData, err := json.Marshal(tokenReq)
		if err != nil {
			return err
		}

		response, err := dev.RequestAuthorization(bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		c := time.Tick(10 * time.Second)
		for _ = range c {
			status, err := dev.TrackAuthorizationProgress(response.Result.TrackID)
			if err != nil {
				return err
			}

			if status.Result.Status == freebox.AuthorizationStatusGranted {
				break
			}
		}

		if err := st.saveAppTokenToFile(response.Result.AppToken); err != nil {
			return err
		}

		if _, err = st.getAppTokenFromFile(); err != nil {
			return err
		}
	}

	return nil
}

func getSessionToken(st *state, dev *freebox.Device, appID string) (*string, error) {
	var sessionToken string

	appTokenBytes, err := st.getAppTokenFromFile()
	if err != nil {
		return nil, err
	}

	sessionTokenBytes, err := st.getSessionTokenFromFile()
	if err != nil {
		challenge, err := dev.GetChallenge()
		if err != nil {
			return nil, err
		}

		session, err := dev.OpenSession(appID, string(appTokenBytes), challenge.Result.Challenge)
		if err != nil {
			return nil, err
		}

		if err := st.saveSessionTokenToFile(session.Result.SessionToken); err != nil {
			return nil, err
		}

		sessionTokenBytes, err = st.getSessionTokenFromFile()
		if err != nil {
			return nil, err
		}
	}
	sessionToken = strings.TrimSpace(string(sessionTokenBytes))

	return &sessionToken, nil
}
