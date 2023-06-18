package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appVersion    string = "development"
	commitHash    string = "?"
	configVersion string = "?"
)

func SetApplicationVersion(app string) {
	appVersion = app
}

func SetCommitHash(commit string) {
	commitHash = commit
}

func SetConfigVersion(config string) {
	configVersion = config
}

type versionOptions struct {
}

func versionCmd(r *rootOptions) *cobra.Command {
	o := &versionOptions{}
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Show version information.",

		Run: func(cmd *cobra.Command, args []string) {
			if !versionCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	return cmd
}

func versionCmdRun(cmd *cobra.Command, r *rootOptions, o *versionOptions, args []string) bool {
	fmt.Printf("application: %s, commit: %s, configuration: %s\n", appVersion, commitHash, configVersion)
	return true
}
