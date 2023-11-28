package main

import "dagger.io/dagger"

// sourceDirectory gives a directory containing the application source code
// excluding non-app directories.
func sourceDirectory(client *dagger.Client) *dagger.Directory {
	return client.Host().Directory(".", dagger.HostDirectoryOpts{
		Exclude: argv("./ci"),
	})
}
