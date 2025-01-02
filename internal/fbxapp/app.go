package fbxapp

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/juli3nk/go-freebox"
)

type App struct {
	id     string
	device *freebox.Device
	state  *state
}

func New(id string, device *freebox.Device, statePath string) (*App, error) {
	app := App{
		id:     id,
		device: device,
	}

	st, err := newState(statePath)
	if err != nil {
		return nil, err
	}

	app.state = st

	return &app, nil
}

func (a *App) Create(name, version, deviceName string) error {
	if _, err := a.state.getAppTokenFromFile(); err != nil {
		tokenReq := freebox.TokenRequest{
			AppID:      a.id,
			AppName:    name,
			AppVersion: version,
			DeviceName: deviceName,
		}

		jsonData, err := json.Marshal(tokenReq)
		if err != nil {
			return err
		}

		response, err := a.device.RequestAuthorization(bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		timeout := time.After(2 * time.Minute) // Set a timeout for safety

		// Label the for loop for breaking out of it
	loop:
		for {
			select {
			case <-ticker.C:
				status, err := a.device.TrackAuthorizationProgress(response.Result.TrackID)
				if err != nil {
					log.Printf("error tracking authorization progress: %v", err)
					continue // Retry on error
				}

				if status.Result.Status == freebox.AuthorizationStatusGranted {
					break loop // Break out of the labeled loop
				}
			case <-timeout:
				return errors.New("authorization timeout") // Timeout occurred
			}
		}

		if err := a.state.saveAppTokenToFile(response.Result.AppToken); err != nil {
			return err
		}

		if _, err = a.state.getAppTokenFromFile(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) GetSessionToken() (*string, error) {
	var sessionToken string

	appTokenBytes, err := a.state.getAppTokenFromFile()
	if err != nil {
		return nil, err
	}

	sessionTokenBytes, err := a.state.getSessionTokenFromFile()
	if err != nil {
		challenge, err := a.device.GetChallenge()
		if err != nil {
			return nil, err
		}

		session, err := a.device.OpenSession(a.id, string(appTokenBytes), challenge.Result.Challenge)
		if err != nil {
			return nil, err
		}

		if err := a.state.saveSessionTokenToFile(session.Result.SessionToken); err != nil {
			return nil, err
		}

		sessionTokenBytes, err = a.state.getSessionTokenFromFile()
		if err != nil {
			return nil, err
		}
	}
	sessionToken = strings.TrimSpace(string(sessionTokenBytes))

	return &sessionToken, nil
}
