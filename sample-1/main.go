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

	golang := client.Container().
		From("golang:1.19").
		WithExec([]string{
			"go", "version",
		})

	version, err := golang.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("golang container stdout: %w", err)
	}

	fmt.Println("Hello from dagger and " + version)

	return nil
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
