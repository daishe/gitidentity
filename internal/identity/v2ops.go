package identity

import (
	"google.golang.org/protobuf/encoding/protojson"

	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
)

func unmarshalConfigV2(cfgBytes []byte) (*configv2.Config, error) {
	cfg := &configv2.Config{}
	if err := (protojson.UnmarshalOptions{AllowPartial: false, DiscardUnknown: false}).Unmarshal(cfgBytes, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
