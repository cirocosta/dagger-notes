package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"dagger.io/dagger"
)

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

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt,
	)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
