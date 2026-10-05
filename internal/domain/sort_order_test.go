package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func TestSortOrder_Next(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		order domain.SortOrder
		want  domain.SortOrder
	}{
		{name: "title moves to updated", order: domain.SortOrderTitle, want: domain.SortOrderUpdated},
		{name: "updated moves to created", order: domain.SortOrderUpdated, want: domain.SortOrderCreated},
		{name: "created wraps around to title", order: domain.SortOrderCreated, want: domain.SortOrderTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.order.Next())
		})
	}
}
