package browse

import "context"

type TrashItemCounter interface {
	CountItems(ctx context.Context) (int, error)
}
