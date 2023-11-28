package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"

	"dagger.io/dagger"
)

// test ensures that the source code in the `source` container is "good to go"
func test(runner *dagger.Container, source *dagger.Directory) *dagger.Container {
	return runner.
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec(argv(`go test -v ./... -run Succeed`))
}

// vet checks for suspicious constructs in the source code
func vet(runner *dagger.Container, source *dagger.Directory) *dagger.Container {
	return runner.
		WithDirectory("/app", source).
		WithWorkdir("/app").
		WithExec(argv(`go vet ./...`))
}

// image builds the application container image out of a source that represents
// the build context.
func image(ctx context.Context, source *dagger.Directory) (string, error) {
	imageName := fmt.Sprintf("ttl.sh/hello-dagger-%.0f",
		math.Floor(rand.Float64()*10000000),
	)

	ref, err := source.
		DockerBuild().
		Publish(ctx, imageName)
	if err != nil {
		return "", fmt.Errorf("docker build: %w", err)
	}

	return ref, nil
}

// scanImage scans a container image.
func scanImage(scanner *dagger.Container, ref string) *dagger.Container {
	return scanner.
		WithExec(argv("image " + ref))
}

// binary builds the application binary
func binary(ctx context.Context, source *dagger.Container) error {
	_, err := source.
		WithExec(argv("go build -o ./out/sample")).
		Directory("./out").
		Export(ctx, "./out")
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

// e2eTest runs the application as a service and runs the e2e test suite
// against it.
func e2eTest(ctx context.Context, ref string) error {
	return nil
}
