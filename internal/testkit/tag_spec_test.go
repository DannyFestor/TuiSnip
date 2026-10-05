package testkit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestTag_Defaults(t *testing.T) {
	t.Parallel()

	tag := testkit.Tag(t, testkit.TagSpec{})

	assert.False(t, tag.ID().IsNil())
	assert.Equal(t, "Tag", tag.Name().String())
	assert.Equal(t, tag.CreatedAt(), tag.UpdatedAt())
}

func TestTag_Spec(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	id := testkit.NewSequentialIDs().NewTagID()

	tag := testkit.Tag(t, testkit.TagSpec{ID: id, Name: "go", CreatedAt: created})

	assert.Equal(t, id, tag.ID())
	assert.Equal(t, "go", tag.Name().String())
	assert.Equal(t, created, tag.UpdatedAt())
}
