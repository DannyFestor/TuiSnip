package look

import "fmt"

const (
	folderNoun  = "Folder"
	snippetNoun = "Snippet"
)

func Counted(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}

	return fmt.Sprintf("%d %ss", count, noun)
}

func FolderAndSnippetCounts(folders, snippets int) []string {
	var counts []string

	if folders > 0 {
		counts = append(counts, Counted(folders, folderNoun))
	}

	if snippets > 0 {
		counts = append(counts, Counted(snippets, snippetNoun))
	}

	return counts
}
