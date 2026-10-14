package exportfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	formatMarker  = "tuisnip-export"
	formatVersion = 1
	indent        = "  "
)

type document struct {
	Format   string          `json:"format"`
	Version  int             `json:"version"`
	Folders  []folderRecord  `json:"folders"`
	Tags     []tagRecord     `json:"tags"`
	Snippets []snippetRecord `json:"snippets"`
}

type header struct {
	Format  string `json:"format"`
	Version int    `json:"version"`
}

func documentOf(library domain.Library) document {
	return document{
		Format:   formatMarker,
		Version:  formatVersion,
		Folders:  recordsOf(library.Folders, folderRecordOf),
		Tags:     recordsOf(library.Tags, tagRecordOf),
		Snippets: recordsOf(library.Snippets, snippetRecordOf),
	}
}

// HTML escaping is off because the file is meant to be read and grepped, and
// Snippet content is full of < and &.
func (d document) encode() ([]byte, error) {
	var encoded bytes.Buffer

	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)

	err := encoder.Encode(d)
	if err != nil {
		return nil, fmt.Errorf("encode export: %w", err)
	}

	return encoded.Bytes(), nil
}

// The header is checked before the strict decode, so a file from another app or a
// newer TuiSnip is refused for what it is rather than for its first unknown field.
func decodeDocument(data []byte) (document, error) {
	var marked header

	err := json.Unmarshal(data, &marked)
	if err != nil {
		return document{}, fmt.Errorf("decode export: %w", err)
	}

	err = marked.requireReadable()
	if err != nil {
		return document{}, err
	}

	return decodeStrictly(data)
}

// An unknown field is refused because a mistyped key, such as "parent-id", would
// otherwise drop its value without a word and import the record changed.
func decodeStrictly(data []byte) (document, error) {
	var decoded document

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&decoded)
	if err != nil {
		return document{}, fmt.Errorf("decode export: %w", err)
	}

	return decoded, nil
}

func (d document) library() (domain.Library, error) {
	folders, foldersErr := entitiesOf(d.Folders, folderRecord.folder)
	tags, tagsErr := entitiesOf(d.Tags, tagRecord.tag)
	snippets, snippetsErr := entitiesOf(d.Snippets, snippetBuilder(tags))

	err := errors.Join(foldersErr, tagsErr, snippetsErr)
	if err != nil {
		return domain.Library{}, err
	}

	return domain.Library{Folders: folders, Tags: tags, Snippets: snippets}, nil
}

func (h header) requireReadable() error {
	switch {
	case h.Format != formatMarker:
		return ErrNotAnExport
	case h.Version > formatVersion:
		return fmt.Errorf("version %d: %w", h.Version, ErrNewerVersion)
	case h.Version != formatVersion:
		return fmt.Errorf("version %d: %w", h.Version, ErrUnknownVersion)
	}

	return nil
}

// An empty list is written as [] rather than null, so a reader never has to
// tell the two apart.
func recordsOf[E, R any](entities []E, recordOf func(entity E) R) []R {
	records := make([]R, 0, len(entities))
	for _, entity := range entities {
		records = append(records, recordOf(entity))
	}

	return records
}

func entitiesOf[R, E any](records []R, entityOf func(record R) (E, error)) ([]E, error) {
	entities := make([]E, 0, len(records))

	var errs []error

	for _, record := range records {
		entity, err := entityOf(record)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		entities = append(entities, entity)
	}

	return entities, errors.Join(errs...)
}
