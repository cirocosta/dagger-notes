package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"dagger.io/dagger"
	"github.com/google/shlex"
	"golang.org/x/sync/errgroup"
)

func test(ctx context.Context, container *dagger.Container) error {
	_, err := container.
		WithExec(argv("go test -v ./... -run Succeed")).
		Stdout(ctx)
	if err != nil {
		return fmt.Errorf("tests: %w", err)
	}

	return nil
}

func vet(ctx context.Context, container *dagger.Container) error {
	_, err := container.
		WithExec(argv("go vet ./...")).
		Stdout(ctx)
	if err != nil {
		return fmt.Errorf("vet: %w", err)
	}

	return nil
}

func binary(ctx context.Context, container *dagger.Container) error {
	_, err := container.
		WithExec(argv("go build -o ./out/sample")).
		Directory("./out").
		Export(ctx, "./out")
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}

	return nil
}

func runPipeline(ctx context.Context, client *dagger.Client) error {
	containerWithSourceCode := client.Container().
		From("golang:1.21").
		WithDirectory("/app", client.Host().Directory("."),
			dagger.ContainerWithDirectoryOpts{
				Exclude: argv("./ci/"),
			}).
		WithWorkdir("/app")

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return test(ctx, containerWithSourceCode)
	})

	g.Go(func() error {
		return vet(ctx, containerWithSourceCode)
	})

	g.Go(func() error {
		return binary(ctx, containerWithSourceCode)
	})

	return g.Wait()
}

func run(ctx context.Context) error {
	client, err := dagger.Connect(ctx,
		dagger.WithLogOutput(os.Stdout),
	)
	if err != nil {
		return fmt.Errorf("dagger connect: %w", err)
	}

	defer client.Close()

	return runPipeline(ctx, client)
}

func argv(s string) []string {
	a, err := shlex.Split(s)
	if err != nil {
		panic(fmt.Errorf("shlex: %w", err))
	}

	return a
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
