package main

import (
	"context"
	"fmt"
	"time"

	"dagger/kaza-presence-tracker-freebox/internal/dagger"

	cplatforms "github.com/containerd/platforms"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// Build container images
func (m *KazaPresenceTrackerFreebox) Build(
	// +optional
	version string,
) (*KazaPresenceTrackerFreebox, error) {
	specifiers := []string{
		"linux/amd64",
		"linux/arm64",
	}
	platforms, err := cplatforms.ParseAll(specifiers)
	if err != nil {
		return nil, err
	}

	git := dag.Git(m.Worktree)

	gitCommit, err := git.GetLatestCommit(context.TODO())
	if err != nil {
		return nil, err
	}

	gitTag, err := git.GetLatestTag(context.TODO())
	if err != nil {
		return nil, err
	}

	gitUncommited, err := git.Uncommited(context.TODO())
	if err != nil {
		return nil, err
	}

	// The binary name
	goBuildPackages := []string{"."}

	appVersion := getVersion(version, gitTag, gitCommit, gitUncommited)

	tsNow := time.Now()

	src := dag.Container().
		WithDirectory("/src", m.Worktree).
		Directory("/src/rootfs")

	binaryPath := fmt.Sprintf("/usr/local/bin/%s", appName)
	entrypoint := "/usr/local/bin/docker-entrypoint.sh"

	for _, platform := range platforms {
		opts := dagger.GoBuildOpts{
			CgoEnabled: "1",
			Musl:       true,
			Arch:       platform.Architecture,
			Os:         platform.OS,
		}
		goBuilder := dag.Go(goVersion, m.Worktree).Build(appName, goBuildPackages, opts)

		image := dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(cplatforms.Format(platform))}).
			From(hassioAddonsBaseImage).
			WithDirectory("/", src).
			WithFile(binaryPath, goBuilder).
			WithUser("nobody:nobody").
			WithEntrypoint([]string{entrypoint}).
			WithLabel("io.hass.name", appName).
			WithLabel("io.hass.description", appDescription).
			WithLabel("io.hass.arch", platform.Architecture).
			WithLabel("io.hass.type", "addon").
			WithLabel("io.hass.version", appVersion).
			WithLabel(specs.AnnotationCreated, tsNow.Format("2006-01-02T15:04:05 -0700")).
			WithLabel(specs.AnnotationSource, fmt.Sprintf("https://%s", appSourceUrl)).
			WithLabel(specs.AnnotationVersion, appVersion).
			WithLabel(specs.AnnotationRevision, gitCommit).
			WithLabel(specs.AnnotationTitle, appName).
			WithLabel(specs.AnnotationDescription, appDescription)

		m.Containers = append(m.Containers, image)
	}

	return m, nil
}
