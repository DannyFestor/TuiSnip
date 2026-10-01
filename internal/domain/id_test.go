package domain_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const storedID = "0192f0c1-7a3b-7c4d-8e5f-6a7b8c9d0e1f"

func TestID_String(t *testing.T) {
	t.Parallel()

	id := domain.SnippetID(uuid.MustParse(storedID))

	assert.Equal(t, storedID, id.String())
}

func TestID_IsNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   domain.FolderID
		want bool
	}{
		{name: "zero value is nil", id: domain.FolderID{}, want: true},
		{name: "nil UUID is nil", id: domain.FolderID(uuid.Nil()), want: true},
		{name: "generated UUID is not nil", id: domain.FolderID(uuid.MustParse(storedID)), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.id.IsNil())
		})
	}
}
