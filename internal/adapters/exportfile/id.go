package exportfile

import (
	"fmt"
	"uuid"
)

type uuidBytes interface {
	~[16]byte
}

func parseID[I uuidBytes](text string) (I, error) {
	parsed, err := uuid.Parse(text)
	if err != nil {
		return I{}, fmt.Errorf("parse id: %w", err)
	}

	return I(parsed), nil
}
