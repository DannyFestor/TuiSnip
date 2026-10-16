package value

import "errors"

type FolderPath struct{ names []FolderName }

func NewFolderPath(raw []string) (FolderPath, error) {
	var (
		names []FolderName
		errs  []error
	)

	for _, name := range raw {
		parsed, err := NewFolderName(name)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		names = append(names, parsed)
	}

	err := errors.Join(errs...)
	if err != nil {
		return FolderPath{}, err
	}

	return FolderPath{names: names}, nil
}

func (p FolderPath) IsEmpty() bool {
	return len(p.names) == 0
}

func (p FolderPath) Strings() []string {
	raw := make([]string, 0, len(p.names))
	for _, name := range p.names {
		raw = append(raw, name.String())
	}

	return raw
}
