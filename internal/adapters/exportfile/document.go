package exportfile

import (
	"bytes"
	"encoding/json"
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

// An empty list is written as [] rather than null, so a reader never has to
// tell the two apart.
func recordsOf[E, R any](entities []E, recordOf func(entity E) R) []R {
	records := make([]R, 0, len(entities))
	for _, entity := range entities {
		records = append(records, recordOf(entity))
	}

	return records
}
