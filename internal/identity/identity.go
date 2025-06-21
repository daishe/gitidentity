package identity

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strings"

	configv2 "github.com/daishe/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/gitinfo"
	"github.com/daishe/gitidentity/internal/runcmd"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
)

func IdentityAsString(i *configv2.Identity) string {
	if id := i.GetIdentifier(); id != "" {
		return id
	}
	n, e := userName(i), userEmail(i)
	if n == "" {
		return fmt.Sprintf("<%s>", e)
	}
	return fmt.Sprintf("%s <%s>", n, e)
}

func IdentitiesAsStrings(is []*configv2.Identity) []string {
	ss := make([]string, len(is))
	for idx, i := range is {
		ss[idx] = IdentityAsString(i)
	}
	return ss
}

func SortIdentities(is []*configv2.Identity) {
	sort.Slice(is, func(i, j int) bool {
		return IdentityAsString(is[i]) < IdentityAsString(is[j])
	})
}

func valueOf(i *configv2.Identity, key string) string {
	v := i.GetValues()
	if v == nil {
		return ""
	}
	return v[key]
}

func userName(i *configv2.Identity) string {
	return valueOf(i, runcmd.GitNameKey)
}

func userEmail(i *configv2.Identity) string {
	return valueOf(i, runcmd.GitEmailKey)
}

func gitNameAndEmailAsIdentity(ctx context.Context, local runcmd.FlagLocalState, global runcmd.FlagGlobalState) (i *configv2.Identity, err error) {
	i = &configv2.Identity{Values: make(map[string]string, 2)}
	i.Values[runcmd.GitNameKey], _, err = runcmd.GetGitConfigValue(ctx, runcmd.GitNameKey, local, global)
	if err != nil {
		return nil, err
	}
	i.Values[runcmd.GitEmailKey], _, err = runcmd.GetGitConfigValue(ctx, runcmd.GitEmailKey, local, global)
	if err != nil {
		return nil, err
	}
	i.Identifier = IdentityAsString(i)
	return i, nil
}

func CurrentIdentity(ctx context.Context, includeGlobal bool) (*configv2.Identity, error) {
	last, has, err := runcmd.GetGitConfigValue(ctx, runcmd.GitLastAppliedKey, runcmd.FlagLocalOn, runcmd.FlagGlobalOff)
	if err != nil {
		return nil, err
	}
	if !has {
		if includeGlobal {
			return gitNameAndEmailAsIdentity(ctx, runcmd.FlagLocalOff, runcmd.FlagGlobalOff)
		}
		return nil, ErrNoCurrentIdentity
	}

	any := &anypb.Any{}
	if err := protojson.Unmarshal([]byte(last), any); err != nil {
		return nil, fmt.Errorf("failed to unmarshall value of %s config key", runcmd.GitLastAppliedKey)
	}
	i, err := unmarshallIdentityFromAny(any)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall value of %s config key: %w", runcmd.GitLastAppliedKey, err)
	}

	for field := range i.GetValues() {
		i.Values[field], _, err = runcmd.GetGitConfigValue(ctx, field, runcmd.FlagLocalState(!includeGlobal), runcmd.FlagGlobalOff)
		if err != nil {
			return nil, err
		}
	}
	return i, nil
}

func GlobalIdentity(ctx context.Context) (*configv2.Identity, error) {
	return gitNameAndEmailAsIdentity(ctx, runcmd.FlagLocalOff, runcmd.FlagGlobalOn)
}

func unsetNameAndEmail(ctx context.Context) error {
	if err := runcmd.SetGitConfigValue(ctx, runcmd.GitNameKey, ""); err != nil {
		return err
	}
	if err := runcmd.SetGitConfigValue(ctx, runcmd.GitEmailKey, ""); err != nil {
		return err
	}
	return nil
}

func UnsetCurrentIdentity(ctx context.Context) error {
	i, err := CurrentIdentity(ctx, false)
	if errors.Is(err, ErrNoCurrentIdentity) {
		return unsetNameAndEmail(ctx)
	}
	if err != nil {
		return err
	}

	for field := range i.GetValues() {
		if err := runcmd.SetGitConfigValue(ctx, field, ""); err != nil {
			return err
		}
	}
	if err := runcmd.SetGitConfigValue(ctx, runcmd.GitLastAppliedKey, ""); err != nil {
		return err
	}
	return unsetNameAndEmail(ctx)
}

func ApplyIdentity(ctx context.Context, i *configv2.Identity) error {
	if err := UnsetCurrentIdentity(ctx); err != nil {
		return err
	}
	if i == nil {
		return nil
	}

	i.Identifier = IdentityAsString(i)
	any, err := marshallIdentityIntoAny(i)
	if err != nil {
		return err
	}
	if err := runcmd.SetGitConfigValue(ctx, runcmd.GitLastAppliedKey, string(any)); err != nil {
		return err
	}
	for key, value := range i.GetValues() {
		if err := runcmd.SetGitConfigValue(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func ApplyIdentityAsArgs(ctx context.Context, i *configv2.Identity) ([]string, error) {
	i.Identifier = IdentityAsString(i)
	any, err := marshallIdentityIntoAny(i)
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, len(i.GetValues())+1)
	args = append(args, fmt.Sprintf("--config=%s=%s", runcmd.GitLastAppliedKey, any))
	for k, v := range i.GetValues() {
		args = append(args, fmt.Sprintf("--config=%s=%s", k, v))
	}
	return args, nil
}

func FirstAutoMatchingIdentity(ctx context.Context, is []*configv2.Identity, info *gitinfo.GitInfo) (*configv2.Identity, error) {
	for _, i := range is {
		matched, err := AutoMatchIdentity(ctx, i, info)
		if err != nil {
			return nil, err
		}
		if matched {
			return i, nil
		}
	}
	return nil, nil //nolint:nilnil // no matching identity found
}

func AutoMatchIdentity(ctx context.Context, i *configv2.Identity, info *gitinfo.GitInfo) (bool, error) {
	for _, ml := range i.GetAutoApplyWhen() {
		verdict, err := matchList(ctx, ml, info)
		if err != nil {
			return false, fmt.Errorf("identity %q: %w", i.GetIdentifier(), err)
		}
		if verdict {
			return true, nil
		}
	}
	return false, nil
}

func matchList(ctx context.Context, ml *configv2.MatchList, info *gitinfo.GitInfo) (bool, error) {
	if len(ml.GetMatch()) == 0 {
		return false, nil
	}
	for _, m := range ml.GetMatch() {
		verdict, err := match(ctx, m, info)
		if err != nil {
			return false, err
		}
		if !verdict {
			return false, nil
		}
	}
	return true, nil
}

func match(ctx context.Context, m *configv2.Match, info *gitinfo.GitInfo) (verdict bool, err error) {
	switch s := m.GetSubject().(type) {
	case *configv2.Match_Env:
		verdict, err = matchEnv(ctx, s.Env)
	case *configv2.Match_Remote:
		verdict, err = matchRemote(ctx, s.Remote, info)
	case *configv2.Match_Command:
		verdict, err = matchCommand(ctx, s.Command)
	case *configv2.Match_ShellScript:
		verdict, err = matchShellScript(ctx, s.ShellScript)
	default:
		return false, nil // unknown matching subject
	}
	return
}

func matchEnv(_ context.Context, m *configv2.MatchEnv) (bool, error) {
	envValue, envExists := os.LookupEnv(m.GetName())
	if !envExists {
		return m.GetTo().GetNegate(), nil
	}
	verdict, err := condition(m.GetTo(), envValue)
	if err != nil {
		return false, fmt.Errorf("matching environment variable %q: %w", m.GetName(), err)
	}
	return verdict, nil
}

func matchRemote(ctx context.Context, m *configv2.MatchRemote, info *gitinfo.GitInfo) (bool, error) {
	for _, r := range info.Remotes {
		ok, err := matchSingleRemote(ctx, m, r)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func matchSingleRemote(_ context.Context, m *configv2.MatchRemote, r *gitinfo.Remote) (bool, error) {
	name, err := condition(m.GetName(), r.Name)
	if err != nil {
		return false, fmt.Errorf("matching remote %q name: %w", r.Name, err)
	}
	url, err := condition(m.GetUrl(), r.Url)
	if err != nil {
		return false, fmt.Errorf("matching remote %q url: %w", r.Name, err)
	}
	return name && url, nil
}

func matchCommand(ctx context.Context, m *configv2.MatchCommand) (bool, error) {
	cmd := make([]string, 0, len(m.GetArgs())+1)
	cmd = append(cmd, m.GetCmd())
	cmd = append(cmd, m.GetArgs()...)
	out, err := runcmd.CommandCombinedOutput(ctx, cmd[0], cmd[1:]...)
	if ee := (&exec.ExitError{}); errors.As(err, &ee) && !m.GetAllowNonZeroExitCode() { // non zero exit code
		return false, nil
	}
	if err != nil {
		return false, runcmd.CommandError(strings.Join(cmd, " "), out, err)
	}
	verdict, err := condition(m.GetOutput(), string(out))
	if err != nil {
		return false, fmt.Errorf("matching command output: %w", err)
	}
	return verdict, nil
}

func matchShellScript(ctx context.Context, m *configv2.MatchShellScript) (bool, error) {
	cmd := append(getAutoMatchShell(), m.GetContent())
	out, err := runcmd.CommandCombinedOutput(ctx, cmd[0], cmd[1:]...)
	if ee := (&exec.ExitError{}); errors.As(err, &ee) && !m.GetAllowNonZeroExitCode() { // non zero exit code
		return false, nil
	}
	if err != nil {
		return false, runcmd.CommandError(strings.Join(cmd, " "), out, err)
	}
	verdict, err := condition(m.GetOutput(), string(out))
	if err != nil {
		return false, fmt.Errorf("matching command output: %w", err)
	}
	return verdict, nil
}

func getAutoMatchShell() []string {
	if runtime.GOOS == "windows" {
		return []string{"powershell.exe", "-NoProfile"}
	}
	return []string{"sh", "-c"}
}

func condition(c *configv2.Condition, target string) (verdict bool, err error) {
	switch c.GetMode() {
	case configv2.ConditionMode_CONTAINS:
		verdict = strings.Contains(target, c.GetValue())
	case configv2.ConditionMode_PREFIX:
		verdict = strings.HasPrefix(target, c.GetValue())
	case configv2.ConditionMode_SUFFIX:
		verdict = strings.HasSuffix(target, c.GetValue())
	case configv2.ConditionMode_FULL:
		verdict = target == c.GetValue()
	case configv2.ConditionMode_SHELL_PATTERN:
		v, err := path.Match(c.GetValue(), target)
		if err != nil {
			return false, fmt.Errorf("matching shell pattern: %w", err)
		}
		verdict = v
	case configv2.ConditionMode_REGEXP:
		r, err := regexp.Compile(c.GetValue())
		if err != nil {
			return false, fmt.Errorf("compiling regexp: %w", err)
		}
		verdict = r.MatchString(target)
	default:
		return false, errors.New("unknown condition mode")
	}

	if c.GetNegate() {
		verdict = !verdict
	}
	return verdict, nil
}
