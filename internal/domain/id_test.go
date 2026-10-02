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

func TestID_Compare(t *testing.T) {
	t.Parallel()

	lower := domain.SnippetID(uuid.MustParse("0194c3a0-0000-7000-8000-000000000001"))
	higher := domain.SnippetID(uuid.MustParse("0194c3a0-0000-7000-8000-0000000000a0"))

	tests := []struct {
		name  string
		id    domain.SnippetID
		other domain.SnippetID
		want  int
	}{
		{name: "orders a lower ID first", id: lower, other: higher, want: -1},
		{name: "orders a higher ID last", id: higher, other: lower, want: 1},
		{name: "finds an ID equal to itself", id: lower, other: lower, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.id.Compare(tt.other))
		})
	}
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
