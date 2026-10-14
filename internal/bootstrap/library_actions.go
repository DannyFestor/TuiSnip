package bootstrap

import (
	"errors"

	"github.com/DannyFestor/TuiSnip/internal/adapters/exportfile"
	"github.com/DannyFestor/TuiSnip/internal/adapters/snippetslab"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
	"github.com/DannyFestor/TuiSnip/internal/adapters/system"
	"github.com/DannyFestor/TuiSnip/internal/app/library"
)

type libraryActions struct {
	export            *library.Export
	importExport      *library.Import
	importSnippetsLab *library.ImportSnippetsLab
}

func newLibraryActions(repository *sqlite.LibraryRepository) (libraryActions, error) {
	export, exportErr := library.NewExport(repository, exportfile.NewWriter())
	importExport, importExportErr := library.NewImport(exportfile.NewReader(), repository)
	importSnippetsLab, importSnippetsLabErr := library.NewImportSnippetsLab(
		snippetslab.NewReader(system.NewIDs(), system.NewClock()),
		repository,
	)

	return libraryActions{export: export, importExport: importExport, importSnippetsLab: importSnippetsLab},
		errors.Join(exportErr, importExportErr, importSnippetsLabErr)
}
