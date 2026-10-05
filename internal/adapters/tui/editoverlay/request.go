package editoverlay

type request int

const (
	requestNothing request = iota
	requestSave
	requestCancel
	requestRefusePasteWithTabs
	requestRefuseOverlongPaste
)
