package runner

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type FDSetHelper func(t *testing.T, target string) *descriptorpb.FileDescriptorSet

var (
	fdsHelper FDSetHelper
	execDir   string
)

func TestMain(m *testing.M) {
	if root, err := filepath.Abs("../"); err != nil {
		log.Fatalf("unable to locate module root: %v", err)
	} else {
		log.Printf("module root: %s", root)
		execDir = root
	}

	if path, err := exec.LookPath("buf"); err == nil {
		log.Println("using: ", path)
		fdsHelper = bufHelper
	} else if path, err = exec.LookPath("protoc"); err == nil {
		log.Println("using: ", path)
		fdsHelper = protocHelper
	} else {
		log.Fatal("could not find buf or protoc in PATH.")
	}

	os.Exit(m.Run())
}

func bufHelper(t *testing.T, target string) *descriptorpb.FileDescriptorSet {
	t.Helper()

	cmd := exec.Command("buf", "build",
		"--path", target,
		"-o", "-",
		"--as-file-descriptor-set",
		"--config", `{"version": "v1"}`,
	)
	cmd.Dir = execDir

	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		require.NoError(t, err, string(exitErr.Stderr))
	} else if err != nil {
		require.NoError(t, err)
	}

	fdset := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(out, fdset)
	require.NoError(t, err, "should be able to unmarshal the fdset")

	return fdset
}

func protocHelper(t *testing.T, target string) *descriptorpb.FileDescriptorSet {
	t.Helper()

	tmp, err := os.CreateTemp("", "protoc-gen-visibility-test")
	require.NoError(t, err, "failed to create temp file")
	defer func() { _ = os.Remove(tmp.Name()) }()

	//nolint:gosec
	cmd := exec.Command("protoc",
		"-o", tmp.Name(),
		"--include_imports",
		target,
	)
	cmd.Dir = execDir

	_, err = cmd.Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		require.NoError(t, err, string(exitErr.Stderr))
	} else if err != nil {
		require.NoError(t, err)
	}

	_, _ = tmp.Seek(0, 0)
	out, err := io.ReadAll(tmp)
	require.NoError(t, err, "should be able to read the raw fdset")

	fdset := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(out, fdset)
	require.NoError(t, err, "should be able to unmarshal the fdset")

	return fdset
}

func initPlugin(t *testing.T, target string) *protogen.Plugin {
	t.Helper()

	cgr := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{target},
		ProtoFile:      fdsHelper(t, target).File,
	}

	p, err := protogen.Options{}.New(cgr)
	require.NoError(t, err)
	return p
}

func TestTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		target  string
		exError bool
	}{
		// dependency protos used by other tests:
		{target: "testdata/protos/deps/no_rules.proto"},
		{target: "testdata/protos/deps/file_public.proto"},
		{target: "testdata/protos/deps/file_subpackages.proto"},
		{target: "testdata/protos/deps/file_package.proto"},
		{target: "testdata/protos/deps/file_private_good_package.proto"},
		{target: "testdata/protos/deps/file_private_good_subpackage.proto"},
		{target: "testdata/protos/deps/subdeps/file_subpackages_import.proto"},
		{target: "testdata/protos/deps/file_invalid_package_name.proto"},

		// these protos are expected to succeed:
		{target: "testdata/protos/good/ignore_transitive_errors.proto"},
		{target: "testdata/protos/good/success_file_private_override.proto"},
		{target: "testdata/protos/good/success_file_private_subpackage.proto"},
		{target: "testdata/protos/good/subgood/success_file_private_subpackage.proto"},

		// these protos are expected to fail:
		{
			target:  "testdata/protos/bad/error_import_package_private.proto",
			exError: true,
		},
		{
			target:  "testdata/protos/bad/error_import_subpackage_private.proto",
			exError: true,
		},
		{
			target:  "testdata/protos/bad/error_import_file_private.proto",
			exError: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.target, func(t *testing.T) {
			t.Parallel()
			p := initPlugin(t, test.target)
			r := New()

			if err := r(p); test.exError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
