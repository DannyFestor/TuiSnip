package snippet

import "github.com/DannyFestor/TuiSnip/internal/domain"

type FragmentInput struct {
	ID       domain.FragmentID
	Title    string
	Language string
	Content  string
}
