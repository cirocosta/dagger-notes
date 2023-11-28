package main

import "dagger.io/dagger"

// scannerContainer gives a container with tooling for scanning images/source
// code.
func scannerContainer(client *dagger.Client) *dagger.Container {
	return client.Container().
		From("ghcr.io/aquasecurity/trivy:canary")
}

// golangContainer gives a container with the golang toolchain ready for
// compiling/testing/etc
func golangContainer(client *dagger.Client) *dagger.Container {
	return client.Container().
		From("golang:1.21")
}
