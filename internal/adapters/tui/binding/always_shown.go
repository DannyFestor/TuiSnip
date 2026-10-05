package binding

import "charm.land/bubbles/v2/key"

func AlwaysShown(hint key.Binding) bool {
	return hint.Help().Desc == labelHelp
}
