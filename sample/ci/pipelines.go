package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"golang.org/x/sync/errgroup"
)

func sync(ctx context.Context, containers []*dagger.Container) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, container := range containers {
		container := container

		g.Go(func() error {
			_, err := container.Sync(ctx)
			return err
		})
	}

	return g.Wait()
}

func runChecks(ctx context.Context, client *dagger.Client) error {
	sourceDir := sourceDirectory(client)
	golangRunner := golangContainer(client)

	if err := sync(ctx,
		[]*dagger.Container{
			test(golangRunner, sourceDir),
			vet(golangRunner, sourceDir),
		},
	); err != nil {
		return err
	}

	return nil
}

func runBuilds(ctx context.Context, client *dagger.Client) error {
	sourceDir := sourceDirectory(client)

	_, err := image(ctx, sourceDir)
	if err != nil {
		return err
	}

	return nil
}

func runPipeline(ctx context.Context, client *dagger.Client) error {
	pipeline := client.Pipeline("sample")

	if err := runChecks(ctx, pipeline.Pipeline("checks")); err != nil {
		return fmt.Errorf("checks: %w", err)
	}

	if err := runBuilds(ctx, pipeline.Pipeline("builds")); err != nil {
		return fmt.Errorf("builds: %w", err)
	}

	return nil
}
