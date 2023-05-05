package main

import (
    "context"
    "fmt"
    "os"

    "dagger.io/dagger"
)

// list of platforms to execute on
var platforms = []dagger.Platform{
    "linux/amd64", // a.k.a. x86_64
    "linux/arm64", // a.k.a. aarch64
}

func main() {
	imageRepo := "docker.io/juli3nk/kaza-presence-tracker:latest"

	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	project := client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: []string{"ci/"},
	})

	platformVariants := make([]*dagger.Container, 0, len(platforms))
	for _, platform := range platforms {
		// initialize this container with the platform
		builder := client.
			Container(dagger.ContainerOpts{Platform: platform}).
			From("golang:1-alpine").

			WithDirectory("/src", project).
			WithWorkdir("/src").

			// mount in an empty dir where the built binary will live
			WithDirectory("/output", client.Directory()).

			WithExec([]string{"apk", "--update", "add", "ca-certificates", "gcc", "git", "musl-dev"}).

			WithExec([]string{"mkdir", "-p", "/output/etc", "/output/etc/ssl/certs"}).

			WithExec([]string{"echo", "nobody:x:65534:"}, dagger.ContainerWithExecOpts{RedirectStdout: "/output/etc/group"}).
			WithExec([]string{"echo", "nobody:x:65534:65534:nobody:/:"}, dagger.ContainerWithExecOpts{RedirectStdout: "/output/etc/passwd"}).
			WithExec([]string{"cp", "/etc/ssl/certs/ca-certificates.crt", "/output/etc/ssl/certs/"}).

			WithExec([]string{"go", "build", "-ldflags", "-linkmode external -extldflags -static -s -w", "-o", "/output/presence-tracker"})

		// select the output directory
		outputDir := builder.Directory("/output")

		// Publish binary on Scratch base
		binaryCtr := client.
			Container(dagger.ContainerOpts{Platform: platform}).
			WithRootfs(outputDir).
			WithUser("nobody:nobody").
			WithEntrypoint([]string{"/presence-tracker"})

		platformVariants = append(platformVariants, binaryCtr)
	}

	// publishing the final image uses the same API as single-platform
	// images, but now additionally specify the `PlatformVariants`
	// option with the containers built before.
	imageDigest, err := client.
		Container().
		Publish(ctx, imageRepo, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("Pushed multi-platform image w/ digest: ", imageDigest)
}
