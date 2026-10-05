package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type scriptCall struct {
	script string
	dir    string
	stdin  string
	env    []string
}

func (c scriptCall) run(t *testing.T) string {
	t.Helper()

	path, err := filepath.Abs(c.script)
	require.NoError(t, err)

	cmd := exec.CommandContext(t.Context(), path) //nolint:gosec // G204: the scripts under test
	cmd.Dir = c.dir
	cmd.Stdin = strings.NewReader(c.stdin)

	cmd.Env = append(os.Environ(), c.env...)

	out, err := cmd.Output()
	require.NoError(t, err)

	return string(out)
}
