package state

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type remembered struct {
	SnippetList snippetList `toml:"snippet_list"`
}

type snippetList struct {
	Sort domain.SortOrder `toml:"sort"`
}

func defaults() remembered {
	return remembered{SnippetList: snippetList{Sort: domain.SortOrderTitle}}
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
