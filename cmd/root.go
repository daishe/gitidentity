package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daishe/gitidentity/internal/logging"
	"github.com/spf13/cobra"
)

type rootOptions struct {
	config    string
	changeDir string
	logging   bool
}

func rootCmd() *cobra.Command {
	o := &rootOptions{}
	cmd := &cobra.Command{
		Use:   "gitidentity",
		Short: "Easily set local git identity",
		Long:  "Gitidentity allows to easily set local git identity.",

		TraverseChildren: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if o.logging {
				logging.Log.SetOutput(cmd.ErrOrStderr())
			}
			if o.changeDir != "" {
				logging.Log.Printf("changing directory to %q", o.changeDir)
				if err := os.Chdir(o.changeDir); err != nil {
					showErr(cmd, err)
					os.Exit(1)
				}
			}
		},
	}

	cmd.PersistentFlags().StringVar(&o.config, "config", defaultConfigPath(), "path to user configuration file")
	cmd.PersistentFlags().BoolVar(&o.logging, "debug", false, "dump debug logs to stderr")
	cmd.PersistentFlags().StringVarP(&o.changeDir, "change-directory", "C", "", "run as if gitidentiry was started in the provided path, instead of the current working directory")

	cmd.AddCommand(addCmd(o))
	cmd.AddCommand(currentCmd(o))
	cmd.AddCommand(setCmd(o))
	cmd.AddCommand(unsetCmd(o))
	cmd.AddCommand(cloneCmd(o))
	cmd.AddCommand(versionCmd(o))
	return cmd
}

func defaultConfigPath() string {
	p, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(p, ".config", "gitidentity", "config.json")
}

func showErr(cmd *cobra.Command, msg interface{}) {
	if msg != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", msg)
	}
}

func Execute(ctx context.Context) error {
	return rootCmd().ExecuteContext(ctx)
}
