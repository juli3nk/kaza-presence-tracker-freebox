package fbxapp

import (
	"os"
	"path/filepath"
)

type state struct {
	path string
}

func newState(path string) (*state, error) {
	return &state{path: path}, nil
}

func (s *state) saveAppTokenToFile(token string) error {
	if err := os.WriteFile(filepath.Join(s.path, "app-token.txt"), []byte(token), 0666); err != nil {
		return err
	}

	return nil
}

func (s *state) getAppTokenFromFile() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(s.path, "app-token.txt"))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *state) saveSessionTokenToFile(token string) error {
	if err := os.WriteFile(filepath.Join(s.path, "session-token.txt"), []byte(token), 0666); err != nil {
		return err
	}

	return nil
}

func (s *state) getSessionTokenFromFile() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(s.path, "session-token.txt"))
	if err != nil {
		return nil, err
	}

	return data, nil
}
