package main

import (
	"dagger/kaza-presence-tracker-freebox/internal/dagger"
)

const (
	appName        = "presence-tracker"
	appSourceUrl   = "github.com/juli3nk/kaza-presence-tracker-freebox"
	appDescription = ""
)

type RegistryAuth struct {
	Address  string
	Username string
	Secret   *dagger.Secret
}

type KazaPresenceTrackerFreebox struct {
	Worktree     *dagger.Directory
	RegistryAuth *RegistryAuth
	Containers   []*dagger.Container
}

func New(
	source *dagger.Directory,
	// +optional
	registryAddress string,
	// +optional
	registryUsername string,
	// +optional
	registrySecret *dagger.Secret,
) *KazaPresenceTrackerFreebox {
	kaza := KazaPresenceTrackerFreebox{Worktree: source}

	if len(registryAddress) > 0 {
		registryAuth := RegistryAuth{
			Address:  registryAddress,
			Username: registryUsername,
			Secret:   registrySecret,
		}

		kaza.RegistryAuth = &registryAuth
	}

	return &kaza
}
