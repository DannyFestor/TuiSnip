package tagrefusal_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagrefusal"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

func TestText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "words a blank name", err: value.ErrBlankTagName, want: "Tag name is blank"},
		{
			name: "words a name with a comma",
			err:  fmt.Errorf("wrapped: %w", value.ErrTagNameHasComma),
			want: "Tag name contains a comma",
		},
		{name: "falls back to the generic failure", err: errors.New("other"), want: look.FailureText},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tagrefusal.Text(tt.err))
		})
	}
}
