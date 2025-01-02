package main

import (
	"context"
	"fmt"

	"dagger/kaza-presence-tracker-freebox/internal/dagger"

	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func (m *KazaPresenceTrackerFreebox) Publish(
	ctx context.Context,
	// +optional
	registryNamespace string,
) error {
	if len(m.Containers) == 0 {
		return fmt.Errorf("error: build containers first")
	}

	appVersion, err := m.Containers[0].Label(ctx, specs.AnnotationVersion)
	if err != nil {
		return err
	}

	ctr := dag.Container()

	imageName := fmt.Sprintf("%s:%s", appName, appVersion)

	if m.RegistryAuth != nil {
		imageName = fmt.Sprintf("%s/%s/%s", m.RegistryAuth.Address, registryNamespace, imageName)

		ctr = ctr.WithRegistryAuth(m.RegistryAuth.Address, m.RegistryAuth.Username, m.RegistryAuth.Secret)
	}

	_, err = ctr.Publish(ctx, imageName, dagger.ContainerPublishOpts{
		PlatformVariants: m.Containers,
	})

	return err
}
