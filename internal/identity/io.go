package identity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/logging"
)

func defaultConfigWritePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gitidentity", "config.yaml"), nil
}

func possibleDefaultConfigPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".config", "gitidentity", "config.json"),
		filepath.Join(home, ".config", "gitidentity", "config.yaml"),
		filepath.Join(home, ".config", "gitidentity", "config.yml"),
	}
}

func possibleConfigPathsInDir(dir string) []string {
	return []string{
		filepath.Join(dir, ".gitidentity.json"),
		filepath.Join(dir, ".gitidentity.yaml"),
		filepath.Join(dir, ".gitidentity.yml"),
	}
}

func extraConfigPaths() []string {
	return strings.Split(os.Getenv("GITIDENTITY_EXTRA_CONFIGS"), string(os.PathListSeparator))
}

type readConfigSetup struct {
	readExtraConfigs  bool
	readParentConfigs bool
}

type ConfigReadOption interface {
	configure(s *readConfigSetup)
}

type configReadOptionFunc func(*readConfigSetup)

func (fn configReadOptionFunc) configure(s *readConfigSetup) { fn(s) }

func WithExtraConfigs(with bool) ConfigReadOption {
	return configReadOptionFunc(func(s *readConfigSetup) {
		s.readExtraConfigs = true
	})
}

func WithParentDirsConfigs(with bool) ConfigReadOption {
	return configReadOptionFunc(func(s *readConfigSetup) {
		s.readParentConfigs = true
	})
}

func ReadConfig(path string, opts ...ConfigReadOption) (*configv2.Config, error) {
	s := &readConfigSetup{
		readExtraConfigs:  true,
		readParentConfigs: true,
	}
	for _, opt := range opts {
		opt.configure(s)
	}

	cfg := EmptyConfig()
	if path != "" {
		_, new, _, err := readConfigFile(path)
		if err != nil {
			return nil, err
		}
		cfg = mergeConfigs(cfg, new)
	}

	if s.readExtraConfigs {
		new, err := readConfig_extraConfigs()
		if err != nil {
			return nil, err
		}
		cfg = mergeConfigs(cfg, new)
	}

	if s.readParentConfigs {
		new, err := readConfig_parentConfigs()
		if err != nil {
			return nil, err
		}
		cfg = mergeConfigs(cfg, new)
	}

	if path == "" {
		_, new, _, err := readConfigFile(possibleDefaultConfigPaths()...)
		if err != nil && !errors.Is(err, ErrConfigFileNotFound) {
			return nil, err
		}
		cfg = mergeConfigs(cfg, new)
	}

	return cfg, nil
}

func ReadConfigForWriting(path string) (string, *configv2.Config, Format, error) {
	paths := []string{path}
	if path == "" {
		paths = possibleDefaultConfigPaths()
	}
	path, cfg, format, err := readConfigFile(paths...)
	if err != nil {
		if errors.Is(err, ErrConfigFileNotFound) {
			path, err := defaultConfigWritePath()
			if err != nil {
				return "", nil, FormatUnknown, err
			}
			return path, EmptyConfig(), FormatYAML, nil
		}
		return "", nil, FormatUnknown, err
	}
	return path, cfg, format, nil
}

func readConfig_extraConfigs() (*configv2.Config, error) {
	merged := EmptyConfig()
	extraPaths := slices.DeleteFunc(extraConfigPaths(), func(path string) bool { return path == "" })
	logging.Log.Printf("loading #%d extra configs, paths %q", len(extraPaths), extraPaths)
	for _, path := range extraPaths {
		_, cfg, _, err := readConfigFile(path)
		if err != nil {
			return nil, err
		}
		merged = mergeConfigs(merged, cfg)
	}
	return merged, nil
}

func readConfig_parentConfigs() (*configv2.Config, error) {
	merged := EmptyConfig()
	dir, err := filepath.Abs(".")
	if err != nil {
		return nil, fmt.Errorf("evaluating path to working directory: %w", err)
	}
	logging.Log.Printf("loading parent dirs configs, starting from current working directory %q", dir)
	for {
		isWorkTree, err := isWithinGitWorkTree(dir)
		if err != nil {
			return nil, err
		}
		if !isWorkTree {
			_, cfg, _, err := readConfigFile(possibleConfigPathsInDir(dir)...)
			if err != nil && !errors.Is(err, ErrConfigFileNotFound) {
				return nil, err
			}
			merged = mergeConfigs(merged, cfg)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return merged, nil
}

var ErrConfigFileNotFound = errors.New("unable to determine the location of the configuration file")

func readConfigBytes(paths ...string) (string, []byte, error) {
	firstErr, firstErrPath := error(nil), ""
	for _, p := range paths {
		cfgBytes, err := os.ReadFile(p)
		if err == nil {
			return p, cfgBytes, nil
		}
		logging.Log.Printf("config %q reading attempt failed: %v", p, err)
		if firstErr == nil && !errors.Is(err, os.ErrNotExist) {
			firstErr, firstErrPath = err, p
		}
	}
	if firstErr == nil { // no paths provided or no config file was found
		return "", nil, ErrConfigFileNotFound
	}
	return firstErrPath, nil, firstErr
}

func readConfigFile(paths ...string) (string, *configv2.Config, Format, error) {
	path, cfgBytes, err := readConfigBytes(paths...)
	if err != nil {
		logging.Log.Printf("config reading failed: attempted paths %q: %v", paths, err)
		return path, nil, FormatUnknown, fmt.Errorf("reading configuration file: %w", err)
	}
	logging.Log.Printf("config %q read", path)
	cfg, format, err := UnmarshalAndValidateConfig(cfgBytes)
	if err != nil {
		logging.Log.Printf("config %q read, unmarshalling or validation failed: %v", path, err)
		return path, nil, format, err
	}
	logging.Log.Printf("config %q read, version %s, #%d numer of entries", path, cfg.GetVersion(), len(cfg.GetList()))
	return path, cfg, format, err
}

func isWithinGitWorkTree(dir string) (bool, error) {
	for {
		is, err := isGitWorkTreeRoot(dir)
		if err != nil {
			return false, fmt.Errorf("checking if a directory %q is located withing git work tree: %w", dir, err)
		}
		if is {
			return true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return false, nil
}

func isGitWorkTreeRoot(dir string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logging.Log.Printf("directory %q is not a git work tree root", dir)
			return false, nil
		}
		logging.Log.Printf("failed checking if directory %q is a git work tree root: %v", dir, err)
		return false, err
	}
	logging.Log.Printf("directory %q is a git work tree root", dir)
	return true, nil
}

func mergeConfigs(x, y *configv2.Config) *configv2.Config {
	switch {
	case x == nil && y == nil:
		return EmptyConfig()
	case x == nil:
		return y
	case y == nil:
		return x
	}

	x.List = append(x.List, y.GetList()...)
	return x
}

func WriteConfig(path string, cfg *configv2.Config, format Format) error {
	logging.Log.Printf("writing config to %q, version %s, #%d numer of entries", path, cfg.GetVersion(), len(cfg.GetList()))
	cfgBytes, err := MarshalConfig(cfg, format)
	if err != nil {
		return fmt.Errorf("marshalling configuration: %w", err)
	}
	if patentDir := filepath.Dir(path); patentDir != "" {
		if err := os.MkdirAll(patentDir, 0o755); err != nil {
			return fmt.Errorf("making directory for user configuration: %w", err)
		}
	}
	if err := osSafeFileWrite(path, cfgBytes, 0o600); err != nil {
		return fmt.Errorf("writing to configuration file: %w", err)
	}
	return nil
}

func osSafeFileWrite(name string, data []byte, perm os.FileMode) error {
	tmpName := fmt.Sprintf("%s.%d.tmp", name, os.Getpid())
	f, err := os.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, name) // atomic on most oses
}
