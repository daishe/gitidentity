package identity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	configv1 "github.com/daishe/gitidentity/config/v1"
	configv2 "github.com/daishe/gitidentity/config/v2"
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
		return nil, fmt.Errorf("version cannot contain whitespace characters")
	} else if v == "" {
		return nil, fmt.Errorf("unset version is unsupported")
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

func UnmarshalAndValidateVersionEntity(p []byte) (VersionEntity, error) {
	ve := &configv2.VersionEntity{}
	if err := (protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}).Unmarshal(p, ve); err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}
	return ve, ValidateVersionEntity(ve)
}

func UnmarshalAndValidateConfig(cfgBytes []byte) (*configv2.Config, error) {
	ev, err := UnmarshalAndValidateVersionEntity(cfgBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	unmarshaller, err := unmarshallerForVersion(ev.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	cfg, err := unmarshaller(cfgBytes)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling configuration: %w", err)
	}
	SortIdentities(cfg.List)
	return cfg, nil
}

func ReadConfig(path string) (*configv2.Config, error) {
	cfgBytes, err := os.ReadFile(path)
	if err != nil {
		logging.Log.Printf("config %q reading failed: %v", path, err)
		return nil, fmt.Errorf("reading configuration file: %w", err)
	}
	logging.Log.Printf("config %q read", path)
	cfg, err := UnmarshalAndValidateConfig(cfgBytes)
	if err != nil {
		logging.Log.Printf("config %q read, unmarshalling or validation failed: %v", path, err)
		return nil, err
	}
	logging.Log.Printf("config %q read, version %s, #%d numer of entries", path, cfg.GetVersion(), len(cfg.GetList()))
	return cfg, err
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

func WriteConfig(path string, cfg *configv2.Config) error {
	logging.Log.Printf("writing config to %q, version %s, #%d numer of entries", path, cfg.GetVersion(), len(cfg.GetList()))
	cfgBytes, err := MarshalConfig(cfg)
	if err != nil {
		return fmt.Errorf("marshalling configuration: %w", err)
	}
	if patentDir := filepath.Dir(path); patentDir != "" {
		if err := os.MkdirAll(patentDir, 0755); err != nil {
			return fmt.Errorf("making directory for user configuration: %w", err)
		}
	}
	if err := osSafeFileWrite(path, cfgBytes, 0600); err != nil {
		return fmt.Errorf("writing to configuration file: %w", err)
	}
	return nil
}

func marshal(m proto.Message) ([]byte, error) {
	return protojson.MarshalOptions{AllowPartial: false, Multiline: true, Indent: "  "}.Marshal(m)
}

func marshalPacked(m proto.Message) ([]byte, error) {
	return protojson.MarshalOptions{AllowPartial: false}.Marshal(m)
}

func MarshalConfig(cfg *configv2.Config) ([]byte, error) {
	return marshal(cfg)
}

func MarshalIdentity(i *configv2.Identity) ([]byte, error) {
	return marshal(i)
}

func marshallIdentityIntoAny(i *configv2.Identity) ([]byte, error) {
	any, err := anypb.New(i)
	if err != nil {
		return nil, err
	}
	return marshalPacked(any)
}

func unmarshallIdentityFromAny(any *anypb.Any) (*configv2.Identity, error) {
	m, err := any.UnmarshalNew()
	if err != nil {
		return nil, err
	}
	switch i := m.(type) {
	case *configv1.Identity:
		return fromIdentityV1ToIdentityV2(i), nil
	case *configv2.Identity:
		return i, nil
	}
	return nil, fmt.Errorf("unknown type %q", any.GetTypeUrl())
}
