package look

const emptyTrashText = "The Trash is empty."

func EmptyTrash(styles Styles) string {
	return styles.Dim.Render(emptyTrashText)
}
