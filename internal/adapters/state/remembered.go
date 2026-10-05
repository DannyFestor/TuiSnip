package state

import (
	"bytes"
	"fmt"
	"uuid"

	"github.com/BurntSushi/toml"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type remembered struct {
	SnippetList snippetList `toml:"snippet_list"`
	Folders     folders     `toml:"folders"`
}

type snippetList struct {
	Sort domain.SortOrder `toml:"sort"`
}

type folders struct {
	Collapsed []uuid.UUID `toml:"collapsed"`
}

func defaults() remembered {
	return remembered{
		SnippetList: snippetList{Sort: domain.SortOrderTitle},
		Folders:     folders{Collapsed: nil},
	}
}

func (f folders) collapsedIDs() []domain.FolderID {
	ids := make([]domain.FolderID, 0, len(f.Collapsed))
	for _, id := range f.Collapsed {
		ids = append(ids, domain.FolderID(id))
	}

	return ids
}

func foldersCollapsing(ids []domain.FolderID) folders {
	collapsed := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		collapsed = append(collapsed, uuid.UUID(id))
	}

	return folders{Collapsed: collapsed}
}

func decode(data []byte) (remembered, error) {
	decoded := defaults()

	_, err := toml.Decode(string(data), &decoded)
	if err != nil {
		return remembered{}, fmt.Errorf("decode state file: %w", err)
	}

	return decoded, nil
}

func (r remembered) encode() ([]byte, error) {
	var encoded bytes.Buffer

	encoder := toml.NewEncoder(&encoded)
	encoder.Indent = ""

	err := encoder.Encode(r)
	if err != nil {
		return nil, fmt.Errorf("encode state file: %w", err)
	}

	return encoded.Bytes(), nil
}
