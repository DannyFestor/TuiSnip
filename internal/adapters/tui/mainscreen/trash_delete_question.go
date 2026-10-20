package mainscreen

import (
	"fmt"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

const (
	deletePermanentlyTitle  = "Delete permanently"
	deletePermanentlyFormat = "Permanently delete %s? This cannot be undone."
	snippetSubjectFormat    = "Snippet %q"
	folderSubjectFormat     = "Folder %q"
	emptyTrashTitle         = "Empty Trash"
	emptyTrashFormat        = "Permanently delete %s in the Trash? This cannot be undone."
	pairLength              = 2
	pairJoint               = " and "
	listSeparator           = ", "
	lastListJoint           = ", and "
	snippetNoun             = "Snippet"
)

type trashTally struct {
	folders  int
	snippets int
}

func deletePermanentlyQuestion(trashItem trash.Item) confirm.Question {
	return confirm.Question{
		Title: deletePermanentlyTitle,
		Text:  fmt.Sprintf(deletePermanentlyFormat, joinedAsList(subjectAndCounts(trashItem))),
	}
}

func subjectAndCounts(trashItem trash.Item) []string {
	if !trashItem.Kind.IsFolder() {
		return []string{fmt.Sprintf(snippetSubjectFormat, trashItem.Name())}
	}

	subject := fmt.Sprintf(folderSubjectFormat, trashItem.Name())
	counts := look.SubfolderAndSnippetCounts(trashItem.Contents.SubfolderCount, trashItem.Contents.SnippetCount)

	return append([]string{subject}, counts...)
}

func joinedAsList(parts []string) string {
	if len(parts) <= pairLength {
		return strings.Join(parts, pairJoint)
	}

	last := len(parts) - 1

	return strings.Join(parts[:last], listSeparator) + lastListJoint + parts[last]
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
	return joinedAsList(look.FolderAndSnippetCounts(t.folders, t.snippets))
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
