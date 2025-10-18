package identity

import (
	"buf.build/go/protoyaml"

	configv1 "github.com/daishe/gitidentity/api/gitidentity/config/v1"
	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
)

func unmarshalConfigV1(cfgBytes []byte) (*configv2.Config, error) {
	cfg := &configv1.Config{}
	if err := (protoyaml.UnmarshalOptions{AllowPartial: false, DiscardUnknown: false}).Unmarshal(cfgBytes, cfg); err != nil {
		return nil, err
	}
	return fromConfigV1ToConfigV2(cfg), nil
}

func fromConfigV1ToConfigV2(v1 *configv1.Config) *configv2.Config {
	v2 := &configv2.Config{
		Version: "v2",
		List:    make([]*configv2.Identity, 0, len(v1.GetList())),
	}
	for _, iv1 := range v1.GetList() {
		v2.List = append(v2.List, fromIdentityV1ToIdentityV2(iv1))
	}
	return v2
}

func fromIdentityV1ToIdentityV2(v1 *configv1.Identity) *configv2.Identity {
	v2 := &configv2.Identity{
		Values: map[string]string{
			"user.name":  v1.GetName(),
			"user.email": v1.GetEmail(),
		},
	}
	v2.Identifier = IdentityAsString(v2)
	return v2
}
