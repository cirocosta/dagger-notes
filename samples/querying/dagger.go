package main

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

func DaggerQuery(ctx context.Context, query string) (any, error) {
	client, err := dagger.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	defer client.Close()

	var data any
	req := &dagger.Request{Query: query}
	resp := &dagger.Response{
		Data: &data,
	}
	if err := client.Do(ctx, req, resp); err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	if len(resp.Errors) != 0 {
		return nil, fmt.Errorf("resp err: %w", resp.Errors)
	}

	return data, nil
}
