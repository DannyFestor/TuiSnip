package snippetslab

import "errors"

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
