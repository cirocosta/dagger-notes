package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/Khan/genqlient/graphql"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := "Basic " + base64.StdEncoding.EncodeToString([]byte(t.key+":"))

	req.Header.Set("Authorization", key)
	return t.wrapped.RoundTrip(req)
}

func CustomQuery(ctx context.Context, query string) (any, error) {
	transport := &authedTransport{
		key:     os.Getenv("DAGGER_SESSION_TOKEN"),
		wrapped: http.DefaultTransport,
	}

	httpClient := http.Client{
		Transport: transport,
	}

	graphqlClient := graphql.NewClient(
		"http://127.0.0.1:8080/query",
		&httpClient,
	)

	response := &graphql.Response{}
	err := graphqlClient.MakeRequest(ctx, &graphql.Request{
		Query:     query,
		Variables: nil,
		OpName:    "",
	}, response)
	if err != nil {
		return nil, fmt.Errorf("gql make req: %w", err)
	}

	return response.Data, nil
}
