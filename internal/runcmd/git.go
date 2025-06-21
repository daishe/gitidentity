package runcmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/daishe/gitidentity/internal/gitinfo"
)

const (
	GitLastAppliedKey         = "gitidentity.lastAppliedIdentity"
	GitNameKey                = "user.name"
	GitEmailKey               = "user.email"
	GitCoreSshCommand         = "core.sshCommand"
	GitCloneDefaultRemoteName = "clone.defaultRemoteName"
)

type FlagLocalState bool

const (
	FlagLocalOn  FlagLocalState = true
	FlagLocalOff FlagLocalState = false
)

type FlagGlobalState bool

const (
	FlagGlobalOn  FlagGlobalState = true
	FlagGlobalOff FlagGlobalState = false
)

var (
	gitExecutable     = "git"
	gitExecutableOnce = sync.Once{}
)

func GitExecutable() string {
	gitExecutableOnce.Do(func() {
		v, ok := os.LookupEnv("GITIDENTITY_GIT_EXECUTABLE")
		if !ok {
			return
		}
		gitExecutable = v
	})
	return gitExecutable
}

func GitNameAndEmail(ctx context.Context, out map[string]string, local FlagLocalState, global FlagGlobalState) (err error) {
	out["user.name"], _, err = GetGitConfigValue(ctx, "user.name", local, global)
	if err != nil {
		return err
	}
	out["user.email"], _, err = GetGitConfigValue(ctx, "user.email", local, global)
	if err != nil {
		return err
	}
	return nil
}

func GetGitConfigValue(ctx context.Context, value string, local FlagLocalState, global FlagGlobalState) (string, bool, error) {
	args := make([]string, 0, 5)
	args = append(args, "config", "--get")
	if local {
		args = append(args, "--local")
	}
	if global {
		args = append(args, "--global")
	}
	args = append(args, value)
	out, err := CommandCombinedOutput(ctx, GitExecutable(), args...)
	if ee := (&exec.ExitError{}); errors.As(err, &ee) && ee.ExitCode() == 1 { // the key is invalid
		return "", false, nil
	}
	if err != nil {
		return "", false, CommandError(fmt.Sprintf("git config %s ...", value), out, err)
	}
	return strings.TrimSpace(string(out)), true, nil
}

func SetGitConfigValue(ctx context.Context, key string, to string) error {
	if to == "" {
		return UnsetGitConfigValue(ctx, key)
	}
	args := []string{"config", "--local", key, to}
	out, err := CommandCombinedOutput(ctx, GitExecutable(), args...)
	if err != nil {
		return CommandError(fmt.Sprintf("git config %s ...", key), out, err)
	}
	return nil
}

func UnsetGitConfigValue(ctx context.Context, key string) error {
	args := []string{"config", "--local", "--unset", key}
	out, err := CommandCombinedOutput(ctx, GitExecutable(), args...)
	if ee := (&exec.ExitError{}); errors.As(err, &ee) && ee.ExitCode() == 5 { // unset of an option which does not exist
		return nil
	}
	if err != nil {
		return CommandError(fmt.Sprintf("git config %s ...", key), out, err)
	}
	return nil
}

func GitInfoFromDir(ctx context.Context) (*gitinfo.GitInfo, error) {
	gi := &gitinfo.GitInfo{}

	remotes, err := listGitRemotes(ctx)
	if err != nil {
		return nil, err
	}
	gi.Remotes = make(gitinfo.Remotes, 0, len(remotes))
	for _, remote := range remotes {
		fields := strings.Fields(remote)
		if len(fields) != 3 {
			continue // skip invalid lines
		}
		gi.Remotes = append(gi.Remotes, &gitinfo.Remote{
			Name: fields[0],
			Url:  fields[1],
		})
	}

	return gi, nil
}

func listGitRemotes(ctx context.Context) ([]string, error) {
	out, err := CommandCombinedOutput(ctx, GitExecutable(), "remote", "-v")
	if err != nil {
		return nil, CommandError("git remote ...", out, err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}
