package editoverlay

import (
	"errors"
	"slices"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func shownFieldErrors(err error) []domain.FieldError {
	fragmentErr, ok := errors.AsType[domain.FragmentError](err)
	if !ok {
		return domain.FieldErrors(err)
	}

	snippetErrors := slices.DeleteFunc(domain.FieldErrors(err), func(fieldErr domain.FieldError) bool {
		return isFragmentField(fieldErr.Field)
	})

	return append(snippetErrors, domain.FieldErrors(fragmentErr)...)
}

func isFragmentField(field domain.Field) bool {
	return slices.Contains([]domain.Field{domain.FieldFragmentTitle, domain.FieldLanguage, domain.FieldContent}, field)
}
