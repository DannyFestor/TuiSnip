package snippetslab

import (
	"fmt"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	labelSeparator = "\n"
	noteSeparator  = "\n\n"
)

type labelledNote struct {
	label string
	text  string
}

func descriptionOf(records []fragmentRecord) (value.Description, error) {
	notes, err := labelledNotes(records)
	if err != nil {
		return value.Description{}, err
	}

	description, err := value.NewDescription(joinedNotes(notes))
	if err != nil {
		return value.Description{}, fmt.Errorf("description: %w", err)
	}

	return description, nil
}

func labelledNotes(records []fragmentRecord) ([]labelledNote, error) {
	notes := make([]labelledNote, 0, len(records))

	for i, record := range records {
		text, err := plainNote(record.Note, record.NoteAttributes)
		if err != nil {
			return nil, fmt.Errorf("fragment %d: note: %w", i+1, err)
		}

		if text != "" {
			notes = append(notes, labelledNote{label: record.label(), text: text})
		}
	}

	return notes, nil
}

func joinedNotes(notes []labelledNote) string {
	if len(notes) == 1 {
		return notes[0].text
	}

	labelled := make([]string, 0, len(notes))
	for _, note := range notes {
		labelled = append(labelled, note.label+labelSeparator+note.text)
	}

	return strings.Join(labelled, noteSeparator)
}
