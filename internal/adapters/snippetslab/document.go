package snippetslab

import (
	"encoding/json"
	"fmt"
	"os"
)

type document struct {
	Contents *contents `json:"contents"`
}

type contents struct {
	Folders  []folderRecord  `json:"folders"`
	Tags     []tagRecord     `json:"tags"`
	Snippets []snippetRecord `json:"snippets"`
}

// Unknown keys are ignored, unlike in a TuiSnip Export, because SnippetsLab writes
// keys its published schema leaves out.
func contentsAt(path string) (contents, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: the user names the file to import
	if err != nil {
		return contents{}, fmt.Errorf("read export: %w", err)
	}

	var decoded document

	err = json.Unmarshal(data, &decoded)
	if err != nil {
		return contents{}, fmt.Errorf("decode export: %w", err)
	}

	if decoded.Contents == nil {
		return contents{}, ErrNotAnExport
	}

	return *decoded.Contents, nil
}
