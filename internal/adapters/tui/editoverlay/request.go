package editoverlay

type request int

const (
	requestNothing request = iota
	requestSave
	requestCancel
	requestEditTags
	requestPickLanguage
	requestRefuseOverlongPaste
	requestExternalEditor
	requestAddFragment
	requestRemoveFragment
	requestRefuseRemovingLastFragment
)
