package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"buf.build/go/protoyaml"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	configv1 "github.com/daishe/gitidentity/api/gitidentity/config/v1"
	configv2 "github.com/daishe/gitidentity/api/gitidentity/config/v2"
)

type Testdata struct {
	t   *testing.T
	dir string
}

func NewTestdata(t *testing.T) *Testdata {
	td := &Testdata{
		t:   t,
		dir: t.TempDir(),
	}
	t.Logf("test directory: %q", td.dir)

	return td
}

func (td *Testdata) FilePath(names ...string) string {
	path := td.dir
	for _, o := range names {
		path = filepath.Join(path, strings.ReplaceAll(o, "/", string(filepath.Separator)))
	}
	return path
}

func (td *Testdata) MustMkdirAll(name string) {
	require.NoError(td.t, os.MkdirAll(td.FilePath(name), 0o775))
}

func (td *Testdata) MustReadFile(name string) []byte {
	data, err := os.ReadFile(td.FilePath(name))
	require.NoError(td.t, err)
	return data
}

func (td *Testdata) MustWriteFile(name string, data []byte) {
	path := td.FilePath(name)
	require.NoError(td.t, os.MkdirAll(filepath.Dir(path), 0o775))
	require.NoError(td.t, os.WriteFile(path, data, 0o664)) //nolint:gosec // This is jut the test and do not need to be secure
}

func (td *Testdata) gitMustExists() {
	_, err := exec.LookPath("git")
	if err != nil {
		td.t.Skipf("failed to locate git executable: %v", err)
	}
}

func (td *Testdata) RunGit(args ...string) ([]byte, error) {
	td.gitMustExists()

	ctx, cancel := context.WithTimeout(td.t.Context(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	return cmd.CombinedOutput()
}

func (td *Testdata) MustRunGit(args ...string) []byte {
	out, err := td.RunGit(args...)
	require.NoError(td.t, err, "args: %v, output: %q", args, out)
	return out
}

func (td *Testdata) RunGitIdentity(args ...string) ([]byte, error) {
	td.gitMustExists()

	root, err := filepath.Abs("../..")
	require.NoError(td.t, err, "cannot get absolute path to project root")

	ctx, cancel := context.WithTimeout(td.t.Context(), 5*time.Second)
	defer cancel()

	a := append([]string{"run", root}, args...)
	td.t.Logf("--- go %v", a)
	cmd := exec.CommandContext(ctx, "go", a...)
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

func (td *Testdata) MustRunGitIdentity(args ...string) []byte {
	out, err := td.RunGitIdentity(args...)
	require.NoError(td.t, err, "args: %v, output: %q", args, out)
	return out
}

func (td *Testdata) RunGitIdentityWithInput(input []byte, args ...string) ([]byte, error) {
	td.gitMustExists()

	root, err := filepath.Abs("../..")
	require.NoError(td.t, err, "cannot get absolute path to project root")

	ctx, cancel := context.WithTimeout(td.t.Context(), 5*time.Second)
	defer cancel()

	a := append([]string{"run", root}, args...)
	cmd := exec.CommandContext(ctx, "go", a...)
	cmd.Env = os.Environ()
	cmd.Stdin = bytes.NewBuffer(input)
	return cmd.CombinedOutput()
}

func (td *Testdata) MustRunGitIdentityWithInput(input []byte, args ...string) []byte {
	out, err := td.RunGitIdentityWithInput(input, args...)
	require.NoError(td.t, err, "args: %v, output: %q", args, out)
	return out
}

func MustMarshalJSON(t *testing.T, m proto.Message) []byte {
	data, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(m)
	require.NoError(t, err)
	return data
}

func MustUnmarshalJSON[M proto.Message](t *testing.T, b []byte, m M) M {
	err := protojson.UnmarshalOptions{AllowPartial: false, DiscardUnknown: false}.Unmarshal(b, m)
	require.NoError(t, err)
	return m
}

func MustMarshalYAML(t *testing.T, m proto.Message) []byte {
	data, err := protoyaml.MarshalOptions{Indent: 2}.Marshal(m)
	require.NoError(t, err)
	return data
}

func MustUnmarshalYAML[M proto.Message](t *testing.T, b []byte, m M) M {
	err := protoyaml.UnmarshalOptions{AllowPartial: false, DiscardUnknown: false}.Unmarshal(b, m)
	require.NoError(t, err)
	return m
}

func Diff(want, got any) string {
	return cmp.Diff(want, got, protocmp.Transform())
}

func ConfigV1(list ...*configv1.Identity) *configv1.Config {
	return &configv1.Config{
		Version: "v1",
		List:    append([]*configv1.Identity(nil), list...),
	}
}

func ConfigV2(list ...*configv2.Identity) *configv2.Config {
	list = slices.Clone(list)
	slices.SortFunc(list, func(i, j *configv2.Identity) int {
		return strings.Compare(i.GetIdentifier(), j.GetIdentifier())
	})
	return &configv2.Config{
		Version: "v2",
		List:    list,
	}
}

func NewIdentityV1() (*configv1.Identity, *configv2.Identity) {
	v2 := NewIdentityV2()
	v1 := &configv1.Identity{
		Name:  v2.GetValues()["user.name"],
		Email: v2.GetValues()["user.email"],
	}
	return v1, v2
}

func NewIdentityV2() *configv2.Identity {
	name := uuid.NewString()
	email := name + "@example.com"
	return &configv2.Identity{
		Identifier: fmt.Sprintf("%s <%s>", name, email),
		Values: map[string]string{
			"user.name":  name,
			"user.email": email,
		},
	}
}
