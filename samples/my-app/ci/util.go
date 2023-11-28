package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/google/shlex"
	"golang.org/x/sync/errgroup"
)

func argv(s string) []string {
	a, err := shlex.Split(s)
	if err != nil {
		panic(fmt.Errorf("shlex: %w", err))
	}

	return a
}

func sync(ctx context.Context, containers ...*dagger.Container) error {
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
