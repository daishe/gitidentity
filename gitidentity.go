package main

import (
	"context"
	"fmt"
	"os"

	"github.com/daishe/gitidentity/cmd"
)

var (
	Version              = "development"
	Commit               = "?"
	ConfigurationVersion = "v1"
)

func main() {
	cmd.SetApplicationVersion(Version)
	cmd.SetCommitHash(Commit)
	cmd.SetConfigVersion(ConfigurationVersion)

	if err := cmd.Execute(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
