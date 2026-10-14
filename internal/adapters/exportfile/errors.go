package exportfile

import "errors"

var (
	ErrNotAnExport    = errors.New("exportfile: not a TuiSnip Export")
	ErrNewerVersion   = errors.New("exportfile: the Export is from a newer TuiSnip; upgrade TuiSnip to import it")
	ErrUnknownVersion = errors.New("exportfile: the Export has a version TuiSnip never wrote")
)
