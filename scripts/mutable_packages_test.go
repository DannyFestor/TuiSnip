package scripts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutablePackages(t *testing.T) {
	t.Parallel()

	module := newModule(t, map[string]string{
		"code/code.go":             "package code\n",
		"testonly/only_test.go":    "package testonly_test\n",
		"feature/doc.go":           "//go:build feature\n\npackage feature\n",
		"e2e/doc.go":               "//go:build e2e\n\npackage e2e\n",
		"testonly/child/child.go":  "package child\n",
		"testonly/child/x_test.go": "package child_test\n",
	})

	tests := []struct {
		name     string
		packages string
		want     string
	}{
		{name: "prints nothing for no packages", packages: "", want: ""},
		{name: "keeps a package with code", packages: "./code\n", want: "./code\n"},
		{name: "drops a package with only test files", packages: "./testonly\n", want: ""},
		{name: "keeps a package whose code the tags include", packages: "./feature\n", want: "./feature\n"},
		{name: "drops a package whose code the tags exclude", packages: "./e2e\n", want: ""},
		{
			name:     "keeps a package with code under a dropped package",
			packages: "./testonly\n./testonly/child\n",
			want:     "./testonly/child\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := scriptCall{
				script: "mutable-packages.sh",
				dir:    module,
				stdin:  tt.packages,
				env:    []string{"MUTATION_TAGS=feature"},
			}.run(t)

			assert.Equal(t, tt.want, got)
		})
	}
}

func newModule(t *testing.T, files map[string]string) string {
	t.Helper()

	module := t.TempDir()
	writeModuleFile(t, module, "go.mod", "module example.com/m\n\ngo 1.27\n")

	for path, content := range files {
		writeModuleFile(t, module, path, content)
	}

	return module
}

func writeModuleFile(t *testing.T, module, path, content string) {
	t.Helper()

	full := filepath.Join(module, path)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o750))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o600))
}
