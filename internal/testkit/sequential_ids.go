package testkit

import (
	"encoding/binary"
	"sync/atomic"
	"time"
	"uuid"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	// The 48-bit millisecond timestamp fills the first six bytes of a UUIDv7.
	timestampShift = 16
	uuidV7Version  = 0x70
	uuidVariant    = 0x80
	variantMask    = 0x3f
)

type SequentialIDs struct {
	counter atomic.Uint64
}

func NewSequentialIDs() *SequentialIDs {
	return &SequentialIDs{counter: atomic.Uint64{}}
}

func (s *SequentialIDs) NewSnippetID() domain.SnippetID {
	return domain.SnippetID(s.next())
}

func (s *SequentialIDs) NewFragmentID() domain.FragmentID {
	return domain.FragmentID(s.next())
}

func (s *SequentialIDs) NewFolderID() domain.FolderID {
	return domain.FolderID(s.next())
}

func (s *SequentialIDs) NewTagID() domain.TagID {
	return domain.TagID(s.next())
}

func (s *SequentialIDs) next() uuid.UUID {
	return sequentialV7(s.counter.Add(1))
}

func sequentialV7(counter uint64) uuid.UUID {
	var id uuid.UUID

	millis := uint64(defaultTime().UnixMilli())
	binary.BigEndian.PutUint64(id[0:8], millis<<timestampShift)
	id[6] = uuidV7Version
	binary.BigEndian.PutUint64(id[8:16], counter)
	id[8] = id[8]&variantMask | uuidVariant

	return id
}

func defaultTime() time.Time {
	return time.Date(2026, time.January, 1, 9, 0, 0, 0, time.UTC)
}
