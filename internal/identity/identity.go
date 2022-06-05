package identity

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"google.golang.org/protobuf/encoding/protojson"

	configv1 "github.com/daishe/gitidentity/config/v1"
)

type VersionEntity interface {
	GetVersion() string
}

func checkVersionString(v string) error {
	if strings.IndexFunc(v, unicode.IsSpace) != -1 {
		return fmt.Errorf("version cannot contain whitespace characters")
	} else if v == "" {
		return fmt.Errorf("unset version is unsupported")
	} else if v != "v1" {
		return fmt.Errorf("version %s is unsupported", v)
	}
	return nil
}

func ValidateVersionEntity(ve VersionEntity) error {
	return checkVersionString(ve.GetVersion())
}

func UnmarshalAndValidateVersionEntity(p []byte) (VersionEntity, error) {
	ve := &configv1.VersionEntity{}
	if err := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(p, ve); err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}
	return ve, checkVersionString(ve.Version)
}

func UnmarshalAndValidateConfig(cfgBytes []byte) (*configv1.Config, error) {
	if _, err := UnmarshalAndValidateVersionEntity(cfgBytes); err != nil {
		return nil, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	cfg := &configv1.Config{}
	if err := (protojson.UnmarshalOptions{AllowPartial: false, DiscardUnknown: false}).Unmarshal(cfgBytes, cfg); err != nil {
		return nil, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	SortIdentities(cfg.List)
	return cfg, nil
}

func ReadConfig(path string) (*configv1.Config, error) {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading configuration file: %w", err)
	}
	return UnmarshalAndValidateConfig(cfgBytes)
}

func EmptyConfig() *configv1.Config {
	return &configv1.Config{Version: "v1"}
}

func WriteConfig(path string, cfg *configv1.Config) error {
	cfgBytes, err := protojson.MarshalOptions{AllowPartial: false, Multiline: true, Indent: "  "}.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling configuration: %w", err)
	}
	if patentDir := filepath.Dir(path); patentDir != "" {
		if err := os.MkdirAll(patentDir, 0755); err != nil {
			return fmt.Errorf("making directory for user configuration: %w", err)
		}
	}
	if err := os.WriteFile(path, cfgBytes, 0600); err != nil {
		return fmt.Errorf("writing to configuration file: %w", err)
	}
	return nil
}

func CurrentIdentity(includeGlobal bool) (*configv1.Identity, error) {
	i := &configv1.Identity{}
	out := []byte(nil)
	err := error(nil)
	if includeGlobal {
		out, err = exec.Command("git", "config", "--default=", "user.name").CombinedOutput()
	} else {
		out, err = exec.Command("git", "config", "--local", "--default=", "user.name").CombinedOutput()
	}
	if err != nil {
		return nil, commandError("git config user.name ...", out, err)
	}
	i.Name = strings.TrimSpace(string(out))

	if includeGlobal {
		out, err = exec.Command("git", "config", "--default=", "user.email").CombinedOutput()
	} else {
		out, err = exec.Command("git", "config", "--local", "--default=", "user.email").CombinedOutput()
	}
	if err != nil {
		return nil, commandError("git config user.email ...", out, err)
	}
	i.Email = strings.TrimSpace(string(out))
	return i, nil
}

func GlobalIdentity() (*configv1.Identity, error) {
	i := &configv1.Identity{}
	out, err := exec.Command("git", "config", "--global", "--default=", "user.name").CombinedOutput()
	if err != nil {
		return nil, commandError("git config user.name ...", out, err)
	}
	i.Name = strings.TrimSpace(string(out))
	out, err = exec.Command("git", "config", "--global", "--default=", "user.email").CombinedOutput()
	if err != nil {
		return nil, commandError("git config user.email ...", out, err)
	}
	i.Email = strings.TrimSpace(string(out))
	return i, nil
}

func ApplyIdentity(i *configv1.Identity) error {
	out := []byte(nil)
	err := error(nil)
	if n := i.GetName(); n != "" {
		out, err = exec.Command("git", "config", "--local", "user.name", n).CombinedOutput()
	} else {
		out, err = exec.Command("git", "config", "--local", "--unset", "user.name").CombinedOutput()
	}
	if err != nil {
		return commandError("git config user.name ...", out, err)
	}

	if e := i.GetEmail(); e != "" {
		out, err = exec.Command("git", "config", "--local", "user.email", e).CombinedOutput()
	} else {
		out, err = exec.Command("git", "config", "--local", "--unset", "user.email").CombinedOutput()
	}
	if err != nil {
		return commandError("git config user.email ...", out, err)
	}
	return nil
}

func commandError(cmd string, out []byte, err error) error {
	if ee := (&exec.ExitError{}); errors.As(err, &ee) {
		err = fmt.Errorf("setting user email: command %q returned non zero exit code %d", cmd, ee.ExitCode())
	} else {
		err = fmt.Errorf("setting user email: command %q unexpected error: %w", cmd, err)
	}
	if len(out) != 0 {
		err = fmt.Errorf("%w, command output:\n%s", err, out)
	}
	return err
}

func IdentityAsString(i *configv1.Identity) string {
	n, e := i.GetName(), i.GetEmail()
	if n == "" {
		return fmt.Sprintf("<%s>", e)
	}
	return fmt.Sprintf("%s <%s>", n, e)
}

func IdentitiesAsStrings(is []*configv1.Identity) []string {
	ss := make([]string, len(is))
	for idx, i := range is {
		ss[idx] = IdentityAsString(i)
	}
	return ss
}

func SortIdentities(is []*configv1.Identity) {
	sort.Slice(is, func(i, j int) bool {
		if is[i].Name != is[j].Name {
			return is[i].Name < is[j].Name
		}
		return is[i].Email < is[j].Email
	})
}
