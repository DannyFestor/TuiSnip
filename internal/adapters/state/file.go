package state

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type File struct {
	path       string
	logger     *slog.Logger
	mutex      sync.RWMutex
	remembered remembered
}

func Load(ctx context.Context, options Options) *File {
	return &File{
		path:       options.Path,
		logger:     options.Logger,
		mutex:      sync.RWMutex{},
		remembered: readOrDefaults(ctx, options),
	}
}

func (f *File) SortOrder() domain.SortOrder {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.remembered.SnippetList.Sort
}

func (f *File) SaveSortOrder(ctx context.Context, order domain.SortOrder) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	next := f.remembered
	next.SnippetList.Sort = order

	err := f.write(ctx, next)
	if err != nil {
		return fmt.Errorf("state.File.SaveSortOrder: %w", err)
	}

	f.remembered = next

	return nil
}

func (f *File) write(ctx context.Context, next remembered) error {
	data, err := next.encode()
	if err != nil {
		return err
	}

	err = atomicfile.Replace(f.path, data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	f.logger.DebugContext(ctx, "state file written", slog.String(keyPath, f.path))

	return nil
}

func readOrDefaults(ctx context.Context, options Options) remembered {
	read, err := readFile(options.Path)
	if err != nil {
		options.Logger.WarnContext(ctx, "state file replaced by defaults",
			slog.String(keyPath, options.Path), slog.Any(keyError, err))

		return defaults()
	}

	return read
}

func readFile(path string) (remembered, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: state.toml from xdg
	if err != nil {
		return remembered{}, fmt.Errorf("read state file: %w", err)
	}

	return decode(data)
}
