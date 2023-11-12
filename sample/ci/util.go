package main

import (
	"fmt"

	"github.com/google/shlex"
)

func argv(s string) []string {
	a, err := shlex.Split(s)
	if err != nil {
		panic(fmt.Errorf("shlex: %w", err))
	}

	return a
}
