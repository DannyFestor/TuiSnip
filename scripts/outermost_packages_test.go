package scripts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutermostPackages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packages string
		want     string
	}{
		{name: "prints nothing for no packages", packages: "", want: ""},
		{
			name:     "keeps unrelated packages",
			packages: "./internal/app/folder\n./internal/domain\n",
			want:     "./internal/app/folder\n./internal/domain\n",
		},
		{
			name:     "drops a package under a listed package",
			packages: "./internal/tui\n./internal/tui/overlay\n./internal/tui/overlay/confirm\n",
			want:     "./internal/tui\n",
		},
		{
			name:     "drops a subpackage listed before its parent",
			packages: "./internal/tui/overlay\n./internal/tui\n",
			want:     "./internal/tui\n",
		},
		{
			name:     "keeps a sibling that shares a name prefix",
			packages: "./internal/app\n./internal/app-tools\n./internal/application\n",
			want:     "./internal/app\n./internal/app-tools\n./internal/application\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := scriptCall{script: "outermost-packages.sh", stdin: tt.packages}.run(t)

			assert.Equal(t, tt.want, got)
		})
	}
}
