package snippet

import "context"

type ClipboardReader interface {
	ReadClipboard(ctx context.Context) (string, error)
}
