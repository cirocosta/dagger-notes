package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func runPipelines(ctx context.Context, client *dagger.Client) error {
	pipeline := client.Pipeline("sample")

	if err := runChecks(ctx, pipeline.Pipeline("checks")); err != nil {
		return fmt.Errorf("checks: %w", err)
	}

	if err := runBuilds(ctx, pipeline.Pipeline("builds")); err != nil {
		return fmt.Errorf("builds: %w", err)
	}

	return nil
}

func runChecks(ctx context.Context, client *dagger.Client) error {
	sourceDir := sourceDirectory(client)
	golangRunner := golangContainer(client)

	if err := sync(ctx,
		test(golangRunner, sourceDir),
		vet(golangRunner, sourceDir),
	); err != nil {
		return err
	}

	return nil
}

func runBuilds(ctx context.Context, client *dagger.Client) error {
	sourceDir := sourceDirectory(client)

	imageRef, err := image(ctx, sourceDir)
	if err != nil {
		return err
	}

	imageScanRunner := scannerContainer(client)
	if err := sync(ctx,
		scanImage(imageScanRunner, imageRef),
	); err != nil {
		return err
	}

	return nil
}
