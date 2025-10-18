package identity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"buf.build/go/protoyaml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	configv1 "github.com/daishe/gitidentity/api/gitidentity/config/v1"
	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/logging"
)

type unmarshaller func([]byte) (*configv2.Config, error)

var versionStringToUnmarshaller = map[string]unmarshaller{
	"v1": unmarshalConfigV1,
	"v2": unmarshalConfigV2,
}

type VersionEntity interface {
	GetVersion() string
}

func unmarshallerForVersion(v string) (unmarshaller, error) {
	if strings.IndexFunc(v, unicode.IsSpace) != -1 {
		return nil, errors.New("version cannot contain whitespace characters")
	} else if v == "" {
		return nil, errors.New("unset version is unsupported")
	}
	unmarshaller := versionStringToUnmarshaller[v]
	if unmarshaller == nil {
		return nil, fmt.Errorf("version %s is unsupported", v)
	}
	return unmarshaller, nil
}

func ValidateVersionEntity(ve VersionEntity) error {
	_, err := unmarshallerForVersion(ve.GetVersion())
	return err
}

func UnmarshalAndValidateVersionEntity(p []byte) (VersionEntity, Format, error) {
	ve := &configv2.VersionEntity{}
	err := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(p, ve)
	if err == nil {
		return ve, FormatJSON, ValidateVersionEntity(ve)
	}
	err = (protoyaml.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(p, ve)
	if err == nil {
		return ve, FormatYAML, ValidateVersionEntity(ve)
	}
	return nil, FormatUnknown, fmt.Errorf("parsing version: %w", err)
}

func UnmarshalAndValidateConfig(cfgBytes []byte) (*configv2.Config, Format, error) {
	ev, format, err := UnmarshalAndValidateVersionEntity(cfgBytes)
	if err != nil {
		return nil, format, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	unmarshaller, err := unmarshallerForVersion(ev.GetVersion())
	if err != nil {
		return nil, format, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	cfg, err := unmarshaller(cfgBytes)
	if err != nil {
		return nil, format, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	SortIdentities(cfg.GetList())
	return cfg, format, nil
}

func defaultConfigPaths() []string {
	p, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(p, ".config", "gitidentity", "config.json"),
		filepath.Join(p, ".config", "gitidentity", "config.yaml"),
		filepath.Join(p, ".config", "gitidentity", "config.yml"),
	}
}

func readConfigBytes(path string) ([]byte, error) {
	tryPaths := []string(nil)
	if path == "" {
		tryPaths = defaultConfigPaths()
	} else {
		tryPaths = append(tryPaths, path)
	}
	firstErr := error(nil)
	for _, p := range tryPaths {
		cfgBytes, err := os.ReadFile(p)
		if err == nil {
			return cfgBytes, nil
		}
		logging.Log.Printf("config %q reading attempt failed: %v", path, err)
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr == nil { // tryPaths is empty
		return nil, errors.New("unable to determine the location of the configuration file")
	}
	return nil, firstErr
}

func ReadConfig(path string) (*configv2.Config, Format, error) {
	cfgBytes, err := readConfigBytes(path)
	if err != nil {
		logging.Log.Printf("config %q reading failed: %v", path, err)
		return nil, FormatUnknown, fmt.Errorf("reading configuration file: %w", err)
	}
	logging.Log.Printf("config %q read", path)
	cfg, format, err := UnmarshalAndValidateConfig(cfgBytes)
	if err != nil {
		logging.Log.Printf("config %q read, unmarshalling or validation failed: %v", path, err)
		return nil, format, err
	}
	logging.Log.Printf("config %q read, version %s, #%d numer of entries", path, cfg.GetVersion(), len(cfg.GetList()))
	return cfg, format, err
}

func EmptyConfig() *configv2.Config {
	return &configv2.Config{Version: "v2"}
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

type Format string

const (
	FormatUnknown Format = "unknown"
	FormatJSON    Format = "json"
	FormatYAML    Format = "yaml"
)

func marshal(m proto.Message, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return protojson.MarshalOptions{AllowPartial: false, Multiline: true, Indent: "  "}.Marshal(m)
	case FormatYAML:
		return protoyaml.MarshalOptions{AllowPartial: false, Indent: 2}.Marshal(m)
	case FormatUnknown:
		break
	}
	return nil, fmt.Errorf("unknown identity marshalling format %q", format)
}

func marshalPacked(m proto.Message) ([]byte, error) {
	return protojson.MarshalOptions{AllowPartial: false}.Marshal(m)
}

func MarshalConfig(cfg *configv2.Config, format Format) ([]byte, error) {
	return marshal(cfg, format)
}

func MarshalIdentity(i *configv2.Identity, format Format) ([]byte, error) {
	return marshal(i, format)
}

func marshallIdentityIntoAny(i *configv2.Identity) ([]byte, error) {
	a, err := anypb.New(i)
	if err != nil {
		return nil, err
	}
	return marshalPacked(a)
}

func unmarshallIdentityFromAny(a *anypb.Any) (*configv2.Identity, error) {
	m, err := a.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	switch i := m.(type) {
	case *configv1.Identity:
		return fromIdentityV1ToIdentityV2(i), nil
	case *configv2.Identity:
		return i, nil
	}
	return nil, fmt.Errorf("unknown type %q", a.GetTypeUrl())
}
