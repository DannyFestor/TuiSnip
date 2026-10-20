package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

const (
	titleSeparator = "; "
	missingLexer   = "no lexer"
)

func snippetsLabReportText(report library.SnippetsLabReport) string {
	lines := slices.Concat(
		[]string{
			"Imported " + countsPhrase(report.Counts) + ", " + pluralised(report.Counts.Fragments, "Fragment"),
			"Put " + pluralised(report.Counts.InTrash.Snippets, "Snippet") + " in the Trash",
		},
		lostLines("Lost a pin", report.Losses.LostPins),
		lostLines("Lost a Gist link", report.Losses.LostGistLinks),
		lostLines("Lost an attachment", report.Losses.LostAttachments),
		fallbackLines(report.Losses),
	)

	return strings.Join(lines, "\n") + "\n"
}

func lostLines(what string, titles []value.Title) []string {
	if len(titles) == 0 {
		return nil
	}

	names := make([]string, 0, len(titles))
	for _, title := range titles {
		names = append(names, title.String())
	}

	return []string{what + ": " + strings.Join(names, titleSeparator)}
}

func fallbackLines(losses domain.ImportLosses) []string {
	if len(losses.FolderFallbacks)+len(losses.FragmentFallbacks) == 0 {
		return nil
	}

	lines := []string{"Fell back to " + value.PlainText().String() + ":"}

	for _, fallback := range losses.FolderFallbacks {
		lines = append(lines, "  Folder "+fallback.Folder.String()+": "+lexerName(fallback.Lexer))
	}

	for _, fallback := range losses.FragmentFallbacks {
		lines = append(lines, "  "+fragmentPlace(fallback)+": "+lexerName(fallback.Lexer))
	}

	return lines
}

func fragmentPlace(fallback domain.FragmentLanguageFallback) string {
	place := fmt.Sprintf("Snippet %s, Fragment %d", fallback.Snippet, fallback.Position)
	if fallback.Fragment.IsSet() {
		place += " (" + fallback.Fragment.String() + ")"
	}

	return place
}

func lexerName(lexer string) string {
	if lexer == "" {
		return missingLexer
	}

	return lexer
}
