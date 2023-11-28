package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/diegosz/gqlformatter"
)

var (
	query  = flag.String("query", "", "graphql query to submit to dagger")
	client = flag.String("client", "dagger", "client to use (dagger|custom)")
)

func runWithContext(ctx context.Context) error {
	data, err := querier()(ctx)
	if err != nil {
		return err
	}

	mustPrintGQL(*query)
	mustPrintJSON(data)

	return nil
}

func mustPrintGQL(query string) {
	q, err := gqlformatter.FormatQuery(query)
	if err != nil {
		panic(fmt.Errorf("gqlfmt format query: %w", err))
	}

	fmt.Println(q)
}

func mustPrintJSON(d any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(d)
}

type queryFn func(ctx context.Context) (any, error)

func querier() queryFn {
	if *query == "" {
		panic(fmt.Errorf("`query` must be set"))
	}

	switch *client {
	case "dagger":
		return func(ctx context.Context) (any, error) {
			return DaggerQuery(ctx, *query)
		}
	case "custom":
		return func(ctx context.Context) (any, error) {
			return CustomQuery(ctx, *query)
		}
	default:
		panic(fmt.Errorf(
			"`client` must be either 'dagger' or 'custom'",
		))
	}
}

func run() error {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(
		context.Background(), os.Interrupt,
	)
	defer cancel()

	return runWithContext(ctx)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
