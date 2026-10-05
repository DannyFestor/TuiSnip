package folder

type PreviewDeleteRepository interface {
	Finder
	SubfolderCounter
	SubtreeSnippetCounter
}
