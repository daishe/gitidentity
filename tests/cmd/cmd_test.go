//go:build !windows

package cmd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
)

func TestAdd(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identityA := NewIdentityV2()
	td.MustWriteFile("config.json", MustMarshalJSON(t, ConfigV2(identityA)))

	identityB := NewIdentityV2()
	identityB.Values["tmp.test"] = "test"
	td.MustRunGitIdentity("--config", td.FilePath("config.json"), "add", "--id", identityB.GetIdentifier(), "--name", identityB.GetValues()["user.name"], "--email", identityB.GetValues()["user.email"], "--value", "tmp.test="+identityB.GetValues()["tmp.test"])

	require.Empty(t, Diff(ConfigV2(identityA, identityB), MustUnmarshalJSON(t, td.MustReadFile("config.json"), &configv2.Config{})))
}

func TestV1CurrentJSON(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identity, identityV2 := NewIdentityV1()
	td.MustWriteFile("config.json", MustMarshalJSON(t, ConfigV1(identity)))

	td.MustMkdirAll("repo")
	td.MustRunGit("-C", td.FilePath("repo"), "init")
	td.MustRunGit("-C", td.FilePath("repo"), "remote", "add", "origin", "ssh://git@example.com/user/example-repo.git")
	searchQuery := []byte(identity.GetName() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "set")

	outputShort := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=short")
	require.Empty(t, Diff(identityV2.GetIdentifier(), strings.TrimSpace(string(outputShort))))

	outputJSON := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=json")
	require.Empty(t, Diff(identityV2, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	outputYAML := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=yaml")
	require.Empty(t, Diff(identityV2, MustUnmarshalYAML(t, outputYAML, &configv2.Identity{})))
}

func TestV1CurrentYAML(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identity, identityV2 := NewIdentityV1()
	td.MustWriteFile("config.yaml", MustMarshalYAML(t, ConfigV1(identity)))

	td.MustMkdirAll("repo")
	td.MustRunGit("-C", td.FilePath("repo"), "init")
	td.MustRunGit("-C", td.FilePath("repo"), "remote", "add", "origin", "ssh://git@example.com/user/example-repo.git")
	searchQuery := []byte(identity.GetName() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "set")

	outputShort := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=short")
	require.Empty(t, Diff(identityV2.GetIdentifier(), strings.TrimSpace(string(outputShort))))

	outputJSON := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=json")
	require.Empty(t, Diff(identityV2, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	outputYAML := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=yaml")
	require.Empty(t, Diff(identityV2, MustUnmarshalYAML(t, outputYAML, &configv2.Identity{})))
}

func TestV2CurrentJSON(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identity := NewIdentityV2()
	td.MustWriteFile("config.json", MustMarshalJSON(t, ConfigV2(identity)))

	td.MustMkdirAll("repo")
	td.MustRunGit("-C", td.FilePath("repo"), "init")
	td.MustRunGit("-C", td.FilePath("repo"), "remote", "add", "origin", "ssh://git@example.com/user/example-repo.git")
	searchQuery := []byte(identity.GetIdentifier() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "set")

	outputShort := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=short")
	require.Empty(t, Diff(identity.GetIdentifier(), strings.TrimSpace(string(outputShort))))

	outputJSON := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=json")
	require.Empty(t, Diff(identity, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	outputYAML := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=yaml")
	require.Empty(t, Diff(identity, MustUnmarshalYAML(t, outputYAML, &configv2.Identity{})))
}

func TestV2CurrentYAML(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identity := NewIdentityV2()
	td.MustWriteFile("config.yaml", MustMarshalYAML(t, ConfigV2(identity)))

	td.MustMkdirAll("repo")
	td.MustRunGit("-C", td.FilePath("repo"), "init")
	td.MustRunGit("-C", td.FilePath("repo"), "remote", "add", "origin", "ssh://git@example.com/user/example-repo.git")
	searchQuery := []byte(identity.GetIdentifier() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "set")

	outputShort := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=short")
	require.Empty(t, Diff(identity.GetIdentifier(), strings.TrimSpace(string(outputShort))))

	outputJSON := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=json")
	require.Empty(t, Diff(identity, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	outputYAML := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.yaml"), "current", "--format=yaml")
	require.Empty(t, Diff(identity, MustUnmarshalYAML(t, outputYAML, &configv2.Identity{})))
}

func TestSet(t *testing.T) {
	t.Parallel()
	td := NewTestdata(t)

	identityA := NewIdentityV2()
	td.MustWriteFile("config.json", MustMarshalJSON(t, ConfigV2(identityA)))

	td.MustMkdirAll("repo")
	td.MustRunGit("-C", td.FilePath("repo"), "init")
	td.MustRunGit("-C", td.FilePath("repo"), "remote", "add", "origin", "ssh://git@example.com/user/example-repo.git")

	searchQuery := []byte(identityA.GetIdentifier() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "set")

	outputJSON := td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=json")
	require.Empty(t, Diff(identityA, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	identityB := NewIdentityV2()
	identityB.AutoApplyWhen = []*configv2.MatchList{
		{
			Match: []*configv2.Match{
				{
					Subject: &configv2.Match_Remote{
						Remote: &configv2.MatchRemote{
							Url: &configv2.Condition{
								Value: "ssh://git@example.com/user/example-repo.git",
							},
						},
					},
				},
			},
		},
	}
	td.MustWriteFile("config.json", MustMarshalJSON(t, ConfigV2(identityA, identityB)))

	searchQuery = []byte("")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "set", "--only-auto")

	outputJSON = td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=json")
	require.Empty(t, Diff(identityB, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))

	searchQuery = []byte(identityA.GetIdentifier() + "\n")
	td.MustRunGitIdentityWithInput(searchQuery, "-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "set", "--no-auto")

	outputJSON = td.MustRunGitIdentity("-C", td.FilePath("repo"), "--config", td.FilePath("config.json"), "current", "--format=json")
	require.Empty(t, Diff(identityA, MustUnmarshalJSON(t, outputJSON, &configv2.Identity{})))
}
