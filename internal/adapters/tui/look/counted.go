package look

import "fmt"

const (
	folderNoun    = "Folder"
	subfolderNoun = "subfolder"
	snippetNoun   = "Snippet"
)

type count struct {
	number int
	noun   string
}

func Counted(number int, noun string) string {
	if number == 1 {
		return fmt.Sprintf("%d %s", number, noun)
	}

	return fmt.Sprintf("%d %ss", number, noun)
}

func FolderAndSnippetCounts(folders, snippets int) []string {
	return nonzeroCounts(count{number: folders, noun: folderNoun}, count{number: snippets, noun: snippetNoun})
}

func SubfolderAndSnippetCounts(subfolders, snippets int) []string {
	return nonzeroCounts(count{number: subfolders, noun: subfolderNoun}, count{number: snippets, noun: snippetNoun})
}

func nonzeroCounts(counts ...count) []string {
	var counted []string

	for _, each := range counts {
		if each.number > 0 {
			counted = append(counted, Counted(each.number, each.noun))
		}
	}

	return counted
}
