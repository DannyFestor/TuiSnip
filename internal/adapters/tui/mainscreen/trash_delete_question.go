package mainscreen

import (
	"fmt"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

const (
	deletePermanentlyTitle = "Delete permanently"
	snippetDeleteFormat    = "Permanently delete Snippet %q? This cannot be undone."
	folderDeleteFormat     = "Permanently delete Folder %q, %s, and %s? This cannot be undone."
	emptyTrashTitle        = "Empty Trash"
	emptyTrashFormat       = "Permanently delete %s in the Trash? This cannot be undone."
	countsJoint            = " and "
	folderNoun             = "Folder"
	subfolderNoun          = "subfolder"
	snippetNoun            = "Snippet"
)

type trashTally struct {
	folders  int
	snippets int
}

func deletePermanentlyQuestion(trashItem trash.Item) confirm.Question {
	text := fmt.Sprintf(snippetDeleteFormat, trashItem.Name())
	if trashItem.Kind.IsFolder() {
		text = fmt.Sprintf(
			folderDeleteFormat,
			trashItem.Name(),
			look.Counted(trashItem.Contents.SubfolderCount, subfolderNoun),
			look.Counted(trashItem.Contents.SnippetCount, snippetNoun),
		)
	}

	return confirm.Question{Title: deletePermanentlyTitle, Text: text}
}

func emptyTrashQuestion(trashItems []trash.Item) confirm.Question {
	return confirm.Question{
		Title: emptyTrashTitle,
		Text:  fmt.Sprintf(emptyTrashFormat, tallyOf(trashItems).String()),
	}
}

func tallyOf(trashItems []trash.Item) trashTally {
	var tally trashTally

	for index := range trashItems {
		tally = tally.with(trashItems[index])
	}

	return tally
}

func (t trashTally) String() string {
	counts := look.NonZeroCounts(
		look.Tally{Count: t.folders, Noun: folderNoun},
		look.Tally{Count: t.snippets, Noun: snippetNoun},
	)

	return strings.Join(counts, countsJoint)
}

func (t trashTally) with(trashItem trash.Item) trashTally {
	if trashItem.Kind.IsFolder() {
		t.folders += 1 + trashItem.Contents.SubfolderCount
		t.snippets += trashItem.Contents.SnippetCount

		return t
	}

	t.snippets++

	return t
}
