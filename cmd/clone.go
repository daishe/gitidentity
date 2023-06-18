package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/gitinfo"
	"github.com/daishe/gitidentity/internal/identity"
	"github.com/daishe/gitidentity/internal/runcmd"
)

type cloneOptions struct {
}

func cloneCmd(r *rootOptions) *cobra.Command {
	o := &cloneOptions{}
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Perform git clone in an identity aware manner",
		Long:  "Perform git clone in an identity aware manner.",

		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			if !cloneCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	return cmd
}

func cloneCmdRun(cmd *cobra.Command, r *rootOptions, o *cloneOptions, args []string) bool {
	cfg, err := identity.ReadConfig(r.config)
	if err != nil {
		showErr(cmd, err)
		return false
	}

	remoteName := cloneCmd_inferRemoteName(cmd.Context(), args)
	gi := &gitinfo.GitInfo{}
	for _, a := range args {
		if _, err := url.Parse(a); err != nil {
			continue
		}
		gi.Remotes = append(gi.Remotes, &gitinfo.Remote{Name: remoteName, Url: a})
	}

	i, err := identity.FirstAutoMatchingIdentity(cmd.Context(), cfg.GetList(), gi)
	if err != nil {
		showErr(cmd, err)
		return false
	}
	if i == nil {
		i, err = selectIdentityPrompt(cmd.Context(), cfg.GetList())
		if err != nil {
			showErr(cmd, err)
			return false
		}
		if i == nil {
			showErr(cmd, fmt.Errorf("no identity selected"))
			return false
		}
	}

	identityAsArgs, err := identity.ApplyIdentityAsArgs(cmd.Context(), i)
	if err != nil {
		showErr(cmd, err)
		return false
	}

	args = append(identityAsArgs, args...)
	args = append([]string{"clone"}, args...)
	if err := runcmd.CommandPipeOutputAndInput(cmd.Context(), runcmd.GitExecutable(), args...); err != nil {
		showErr(cmd, runcmd.CommandError(fmt.Sprintf("%s %s", runcmd.GitExecutable(), strings.Join(args, " ")), nil, err))
		return false
	}
	return true
}

func cloneCmd_inferRemoteName(ctx context.Context, args []string) string {
	name := "origin"

	fromConfig, _, err := runcmd.GetGitConfigValue(ctx, runcmd.GitCloneDefaultRemoteName, runcmd.FlagLocalOff, runcmd.FlagGlobalOff)
	if fromConfig != "" && err == nil {
		name = fromConfig
	}

	for i := range args {
		value, exists := cloneCmd_inferRemoteNameFromArg(args, i)
		if !exists {
			continue
		}
		name = value
	}

	return name
}

func cloneCmd_inferRemoteNameFromArg(args []string, idx int) (string, bool) {
	if configArg, exists := recoverArgValue("config", "c", args, idx); exists {
		key, value, found := strings.Cut(configArg, "=")
		if found && key != runcmd.GitCloneDefaultRemoteName {
			return value, true
		}
	}
	if originArg, exists := recoverArgValue("origin", "o", args, idx); exists {
		return originArg, true
	}
	return "", false
}

func recoverArgValue(name string, shorthand string, args []string, idx int) (string, bool) {
	if idx < 0 || idx >= len(args) {
		return "", false // argument does not exists
	}
	arg := args[idx]

	arg, match := strings.CutPrefix(arg, "-")
	if !match || len(arg) == 0 {
		return "", false // not a flag
	}

	// full (non shorthand packed) flag
	if tail, match := strings.CutPrefix(arg, "-"); match {
		if tail, match = strings.CutPrefix(tail, name); !match {
			return "", false // name does not match
		}
		if len(tail) != 0 {
			if tail[0] != '=' {
				return "", false // name do not match fully - just a prefix match
			}
			// name match full and value is bundled
			return tail[1:], true
		}
		// value must be in the next argument
		idx += 1
		if idx < 0 || idx >= len(args) {
			return "", false // next argument do not exists
		}
		return args[idx], true // value in the next argument
	}

	// packed flag (one or more shortcuts put together)
	pack, value, hasValue := strings.Cut(arg, "=")
	if len(pack) == 0 {
		return "", false // not a flag
	}
	if !strings.HasSuffix(pack, shorthand) {
		return "", false // packed (shorthand) flag not at the last position in pack (or not at all in pack)
	}
	// packed (shorthand) flag not is at the last position in pack
	if hasValue {
		return value, true // value is bundled with pack
	}
	// value must be in the next argument
	idx += 1
	if idx < 0 || idx >= len(args) {
		return "", false // next argument do not exists
	}
	return args[idx], true // value in the next argument
}
