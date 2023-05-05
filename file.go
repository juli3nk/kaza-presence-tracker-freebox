package main

import (
	"os"
	"path/filepath"
)

const appRootDir = "/srv/presence-tracker/state"

func saveAppTokenToFile(token string) error {
	if err := os.WriteFile(filepath.Join(appRootDir, "app-token.txt"), []byte(token), 0666); err != nil {
		return err
	}

	return nil
}

func getAppTokenFromFile() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(appRootDir, "app-token.txt"))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func saveSessionTokenToFile(token string) error {
	if err := os.WriteFile(filepath.Join(appRootDir, "session-token.txt"), []byte(token), 0666); err != nil {
		return err
	}

	return nil
}

func getSessionTokenFromFile() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(appRootDir, "session-token.txt"))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func removeSessionTokenFile() error {
	if err := os.Remove(filepath.Join(appRootDir, "session-token.txt")); err != nil {
		return err
	}

	return nil
}
