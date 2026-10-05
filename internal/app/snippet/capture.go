package snippet

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Capture struct {
	clipboard ClipboardReader
}

func NewCapture(clipboard ClipboardReader) (*Capture, error) {
	err := domain.RequireDependency("clipboard", clipboard)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewCapture: %w", err)
	}

	return &Capture{clipboard: clipboard}, nil
}

func (c *Capture) Run(ctx context.Context, _ CaptureInput) (value.Content, error) {
	read, err := c.clipboard.ReadClipboard(ctx)
	if err != nil {
		return value.Content{}, fmt.Errorf("snippet.Capture: %w", err)
	}

	if read == "" {
		return value.Content{}, fmt.Errorf("snippet.Capture: %w", domain.ErrClipboardEmpty)
	}

	content, err := value.NewContent(read)
	if err != nil {
		return value.Content{}, fmt.Errorf("snippet.Capture: %w", err)
	}

	return content, nil
}
