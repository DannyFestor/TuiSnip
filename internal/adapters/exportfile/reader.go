package exportfile

import (
	"context"
	"fmt"
	"os"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Reader struct{}

func NewReader() *Reader {
	return &Reader{}
}

func (*Reader) Read(_ context.Context, path string) (domain.Library, error) {
	library, err := libraryAt(path)
	if err != nil {
		return domain.Library{}, fmt.Errorf("exportfile.Reader.Read: %w", err)
	}

	return library, nil
}

func libraryAt(path string) (domain.Library, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: the user names the file to import
	if err != nil {
		return domain.Library{}, fmt.Errorf("read export: %w", err)
	}

	decoded, err := decodeDocument(data)
	if err != nil {
		return domain.Library{}, err
	}

	return decoded.library()
}
