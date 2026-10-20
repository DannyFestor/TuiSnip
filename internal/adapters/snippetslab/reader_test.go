package snippetslab_test

import (
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/snippetslab"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

const (
	exportName  = "snippetslab.json"
	snippetUUID = "6F1B0C52-0000-4000-8000-000000000001"
	created     = "2024-05-01T09:30:00Z"
	modified    = "2024-05-02T10:00:00+09:00"
)

type record = map[string]any

func TestReader_Read(t *testing.T) {
	t.Parallel()

	t.Run("reads a live Snippet at the Root with its Fragments in order and their dates", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"fragments": []record{
				fragmentRecord(record{"title": "main.go", "content": "package main\n"}),
				fragmentRecord(record{"title": "go.mod", "content": "module x\n"}),
			}}),
		}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 1)
		snippet := got.Snippets[0]
		assert.Equal(t, "loop", snippet.Title().String())
		assert.True(t, snippet.AtRoot())
		assert.Equal(t, time.Date(2024, time.May, 1, 9, 30, 0, 0, time.UTC), snippet.CreatedAt())
		assert.Equal(t, time.Date(2024, time.May, 2, 1, 0, 0, 0, time.UTC), snippet.UpdatedAt())
		require.Len(t, snippet.Fragments(), 2)
		assert.Equal(t, "main.go", snippet.Fragments()[0].Title().String())
		assert.Equal(t, "package main\n", snippet.Fragments()[0].Content().String())
		assert.Equal(t, "Go", snippet.Fragments()[0].Language().String())
		assert.Equal(t, "go.mod", snippet.Fragments()[1].Title().String())
		assert.Equal(t, time.Date(2024, time.May, 1, 9, 30, 0, 0, time.UTC), snippet.Fragments()[1].CreatedAt())
	})

	t.Run("drops a Fragment title that is exactly SnippetsLab's default and keeps any other", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"fragments": []record{
				fragmentRecord(record{"title": "Fragment"}),
				fragmentRecord(record{"title": "Fragment 2"}),
				fragmentRecord(record{"title": "fragment"}),
			}}),
		}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 1)
		assert.Equal(t, []string{"", "Fragment 2", "fragment"}, fragmentTitles(got.Snippets[0]))
	})

	t.Run("dates a Snippet by its newest Fragment when that is newer than the Snippet", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"dateModified": created, "fragments": []record{
				fragmentRecord(record{"dateModified": "2025-01-01T00:00:00Z"}),
				fragmentRecord(record{"dateModified": "2024-12-31T00:00:00Z"}),
			}}),
		}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 1)
		assert.Equal(t, time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC), got.Snippets[0].UpdatedAt())
	})

	t.Run("puts each Snippet in SnippetsLab's Trash in the Trash, deleted at its dateDeleted", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"uuid": "live", "title": "live"}),
			snippetRecord(record{"uuid": "deleted", "title": "deleted", "dateDeleted": "2024-06-01T09:00:00+09:00"}),
		}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 2)
		deleted := got.Snippets[1]
		assert.Equal(t, "deleted", deleted.Title().String())
		assert.Equal(t, map[domain.SnippetID]domain.TrashMark{
			deleted.ID(): testkit.TrashMark(t, testkit.TrashMarkSpec{
				Item: deleted, DeletedAt: time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC),
			}),
		}, got.Trash.Snippets)
		assert.Empty(t, got.Trash.Folders)
	})

	t.Run("names each pinned Snippet as having lost its pin", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"title": "pinned", "pinned": true}),
			snippetRecord(record{"title": "unpinned"}),
			snippetRecord(record{"title": "pinned in the Trash", "pinned": true, "dateDeleted": modified}),
		}))

		_, losses := read(t, path)

		assert.Equal(t, []value.Title{title(t, "pinned"), title(t, "pinned in the Trash")}, losses.LostPins)
	})

	t.Run("names each Snippet linked to a Gist as having lost its Gist link", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"title": "shared", "gistIdentifier": "0123456789abcdef0123456789abcdef"}),
			snippetRecord(record{"title": "unshared", "gistIdentifier": nil}),
			snippetRecord(record{"title": "never shared"}),
			snippetRecord(record{
				"title":          "shared, then deleted",
				"gistIdentifier": "fedcba9876543210fedcba9876543210",
				"dateDeleted":    modified,
			}),
		}))

		_, losses := read(t, path)

		assert.Equal(t, []value.Title{title(t, "shared"), title(t, "shared, then deleted")}, losses.LostGistLinks)
	})

	t.Run("names each Snippet with an attachment once as having lost its attachments", func(t *testing.T) {
		t.Parallel()

		attached := fragmentRecord(record{"attachments": []record{attachmentRecord(), attachmentRecord()}})
		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"title": "screenshots", "fragments": []record{
				fragmentRecord(nil), attached, attached,
			}}),
			snippetRecord(record{"title": "plain", "fragments": []record{
				fragmentRecord(record{"attachments": []record{}}),
			}}),
		}))

		_, losses := read(t, path)

		assert.Equal(t, []value.Title{title(t, "screenshots")}, losses.LostAttachments)
	})

	lexers := []struct {
		lexer string
		want  string
	}{
		{lexer: "PhpLexer", want: "PHP"},
		{lexer: "GoLexer", want: "Go"},
		{lexer: "JavascriptLexer", want: "JavaScript"},
		{lexer: "SqlLexer", want: "SQL"},
		{lexer: "DartLexer", want: "Dart"},
		{lexer: "HtmlLexer", want: "HTML"},
		{lexer: "MarkdownLexer", want: "markdown"},
		{lexer: "PostgresLexer", want: "PostgreSQL SQL dialect"},
		{lexer: "TypeScriptLexer", want: "TypeScript"},
		{lexer: "YamlLexer", want: "YAML"},
		{lexer: "DockerLexer", want: "Docker"},
		{lexer: "BashLexer", want: "Bash"},
		{lexer: "CSharpLexer", want: "C#"},
		{lexer: "HtmlPhpLexer", want: "PHTML"},
		{lexer: "JavascriptPhpLexer", want: "JavaScript"},
		{lexer: "TextLexer", want: "plaintext"},
		{lexer: "CssLexer", want: "CSS"},
		{lexer: "MySqlLexer", want: "MySQL"},
		{lexer: "JsonLexer", want: "JSON"},
		{lexer: "TOMLLexer", want: "TOML"},
		{lexer: "ZigLexer", want: "Zig"},
		{lexer: "CLexer", want: "C"},
		{lexer: "CppLexer", want: "C++"},
		{lexer: "JavaLexer", want: "Java"},
		{lexer: "KotlinLexer", want: "Kotlin"},
		{lexer: "OcamlLexer", want: "OCaml"},
		{lexer: "OdinLexer", want: "Odin"},
		{lexer: "RustLexer", want: "Rust"},
	}
	for _, tt := range lexers {
		t.Run("maps "+tt.lexer+" to "+tt.want, func(t *testing.T) {
			t.Parallel()

			path := exportFile(t, contents(nil, nil, []record{
				snippetRecord(record{"fragments": []record{fragmentRecord(record{"language": tt.lexer})}}),
			}))

			got, losses := read(t, path)

			require.Len(t, got.Snippets, 1)
			assert.Equal(t, tt.want, got.Snippets[0].FirstFragment().Language().String())
			assert.Empty(t, losses.FragmentFallbacks)
		})
	}

	t.Run("falls back to plain text for a lexer the table lacks and names the Fragment", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{
			snippetRecord(record{"title": "parser", "fragments": []record{
				fragmentRecord(nil),
				fragmentRecord(record{"title": "grammar.y", "language": "YaccLexer"}),
			}}),
		}))

		got, losses := read(t, path)

		require.Len(t, got.Snippets, 1)
		assert.Equal(t, "plaintext", got.Snippets[0].Fragments()[1].Language().String())
		assert.Equal(t, []domain.FragmentLanguageFallback{{
			Snippet:  title(t, "parser"),
			Position: 2,
			Fragment: fragmentTitle(t, "grammar.y"),
			Lexer:    "YaccLexer",
		}}, losses.FragmentFallbacks)
	})
}

func TestReader_Read_folders(t *testing.T) {
	t.Parallel()

	t.Run("keeps the Folder tree, dated at the Import, and files each Snippet in its Folder", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents([]record{
			folderRecord("go", "GoLexer", folderRecord("cli", "BashLexer")),
		}, nil, []record{
			snippetRecord(record{"title": "flags", "folder": folderUUID("cli")}),
		}))

		got, _ := read(t, path)

		goFolder, cli := folderNamed(t, got, "go"), folderNamed(t, got, "cli")
		assert.True(t, goFolder.AtRoot())
		assert.Equal(t, goFolder.ID(), cli.ParentID())
		assert.Equal(t, "Bash", cli.DefaultLanguage().String())
		assert.Equal(t, importedAt(), cli.CreatedAt())
		assert.Equal(t, importedAt(), cli.UpdatedAt())
		require.Len(t, got.Snippets, 1)
		assert.Equal(t, cli.ID(), got.Snippets[0].FolderID())
	})

	t.Run("gives a Folder without a Default Language its parent's, or plain text at the top", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents([]record{
			folderRecord("go", "GoLexer", folderRecord("cli", "", folderRecord("flags", ""))),
			folderRecord("misc", ""),
		}, nil, nil))

		got, losses := read(t, path)

		assert.Equal(t, "Go", folderNamed(t, got, "cli").DefaultLanguage().String())
		assert.Equal(t, "Go", folderNamed(t, got, "flags").DefaultLanguage().String())
		assert.Equal(t, "plaintext", folderNamed(t, got, "misc").DefaultLanguage().String())
		assert.Empty(t, losses.FolderFallbacks)
	})

	t.Run("falls back to plain text for a Default Language the table lacks and names the Folder", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents([]record{
			folderRecord("go", "GoLexer", folderRecord("grammars", "YaccLexer", folderRecord("lalr", ""))),
		}, nil, nil))

		got, losses := read(t, path)

		assert.Equal(t, "plaintext", folderNamed(t, got, "grammars").DefaultLanguage().String())
		assert.Equal(t, "plaintext", folderNamed(t, got, "lalr").DefaultLanguage().String())
		assert.Equal(t, []domain.FolderLanguageFallback{
			{Folder: folderNamed(t, got, "grammars").Name(), Lexer: "YaccLexer"},
		}, losses.FolderFallbacks)
	})

	t.Run("files a Snippet at the Root when its Folder isn't in the file", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{snippetRecord(record{"folder": folderUUID("gone")})}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 1)
		assert.True(t, got.Snippets[0].AtRoot())
	})

	t.Run("keeps a trashed Snippet in its Folder, or at the Root when the file lacks it", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents([]record{folderRecord("go", "GoLexer")}, nil, []record{
			snippetRecord(record{"title": "filed", "folder": folderUUID("go"), "dateDeleted": modified}),
			snippetRecord(record{"title": "orphaned", "folder": folderUUID("gone"), "dateDeleted": modified}),
		}))

		got, _ := read(t, path)

		require.Len(t, got.Snippets, 2)
		filed, orphaned := got.Snippets[0], got.Snippets[1]
		assert.Equal(t, folderNamed(t, got, "go").ID(), filed.FolderID())
		assert.True(t, orphaned.AtRoot())
		assert.True(t, got.Trash.Snippets[filed.ID()].FormerPath().IsEmpty())
		assert.True(t, got.Trash.Snippets[orphaned.ID()].FormerPath().IsEmpty())
		assert.Len(t, got.Trash.Snippets, 2)
	})
}

func TestReader_Read_tags(t *testing.T) {
	t.Parallel()

	t.Run(
		"reads every Tag, dated at the Import, and drops a Tag a Snippet names but the file lacks",
		func(t *testing.T) {
			t.Parallel()

			path := exportFile(t, contents(nil, []record{
				tagRecord("Go"),
				tagRecord("unused"),
			}, []record{
				snippetRecord(record{"tags": []string{tagUUID("Go"), tagUUID("gone")}}),
			}))

			got, _ := read(t, path)

			require.Len(t, got.Tags, 2)
			assert.Equal(t, "Go", got.Tags[0].Name().String())
			assert.Equal(t, importedAt(), got.Tags[0].CreatedAt())
			assert.Equal(t, "unused", got.Tags[1].Name().String())
			require.Len(t, got.Snippets, 1)
			assert.Equal(t, got.Tags[:1], got.Snippets[0].Tags())
		},
	)
}

func TestReader_Read_notes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fragments []record
		want      string
	}{
		{
			name:      "makes a Snippet's one note its Description",
			fragments: []record{fragmentRecord(record{"note": "Run it with -v."}), fragmentRecord(nil)},
			want:      "Run it with -v.",
		},
		{
			name: "joins several notes in Fragment order, each under its Fragment title or Language",
			fragments: []record{
				fragmentRecord(record{"title": "main.go", "note": "The entry point."}),
				fragmentRecord(record{"note": "  "}),
				fragmentRecord(record{"title": "Fragment", "language": "BashLexer", "note": "Build it."}),
			},
			want: "main.go\nThe entry point.\n\nBash\nBuild it.",
		},
		{
			name:      "counts a note of only whitespace as no note",
			fragments: []record{fragmentRecord(record{"note": " \n\t"})},
			want:      "",
		},
		{
			name: "keeps a link as its text and URL, counting the range in UTF-16 units",
			fragments: []record{fragmentRecord(record{
				"note": "😀 See docs and https://go.dev here.",
				"noteAttributes": []record{
					{"name": "strong", "range": []int{0, 2}},
					{"name": "link", "range": []int{7, 4}, "stringValue": "https://pkg.go.dev"},
					{"name": "link", "range": []int{16, 14}, "stringValue": "https://go.dev"},
				},
			})},
			want: "😀 See docs (https://pkg.go.dev) and https://go.dev here.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := exportFile(t, contents(nil, nil, []record{snippetRecord(record{"fragments": tt.fragments})}))

			got, _ := read(t, path)

			require.Len(t, got.Snippets, 1)
			assert.Equal(t, tt.want, got.Snippets[0].Description().String())
		})
	}
}

func TestReader_Read_refusals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		document    record
		wantErr     error
		wantMessage []string
	}{
		{
			name:        "refuses JSON that isn't a SnippetsLab export",
			document:    record{"format": "tuisnip-export", "version": 1},
			wantErr:     snippetslab.ErrNotAnExport,
			wantMessage: nil,
		},
		{
			name: "refuses a Snippet title over 200 characters and names the Snippet",
			document: contents(nil, nil, []record{
				snippetRecord(record{"title": strings.Repeat("a", 201)}),
			}),
			wantErr:     value.ErrTitleTooLong,
			wantMessage: []string{"snippet " + snippetUUID + ": "},
		},
		{
			name: "refuses a Fragment updated before it was created and names it by position",
			document: contents(nil, nil, []record{snippetRecord(record{"fragments": []record{
				fragmentRecord(nil),
				fragmentRecord(record{"dateCreated": modified, "dateModified": created}),
			}})}),
			wantErr:     domain.ErrUpdatedBeforeCreate,
			wantMessage: []string{"snippet " + snippetUUID + ": fragment 2: "},
		},
		{
			name:        "refuses a Snippet deleted before it was created and names the Snippet",
			document:    contents(nil, nil, []record{snippetRecord(record{"dateDeleted": "2024-04-30T00:00:00Z"})}),
			wantErr:     domain.ErrDeletedBeforeCreate,
			wantMessage: []string{"snippet " + snippetUUID + ": "},
		},
		{
			name: "refuses a note link that runs past its note",
			document: contents(nil, nil, []record{snippetRecord(record{"fragments": []record{
				fragmentRecord(record{"note": "docs", "noteAttributes": []record{
					{"name": "link", "range": []int{2, 5}, "stringValue": "https://go.dev"},
				}}),
			}})}),
			wantErr:     snippetslab.ErrMalformedNoteLink,
			wantMessage: []string{"snippet " + snippetUUID + ": fragment 1: note: "},
		},
		{
			name: "names a bad Folder inside a bad Folder",
			document: contents(
				[]record{folderRecord(" ", "", folderRecord("  ", ""))},
				nil,
				nil,
			),
			wantErr:     value.ErrBlankFolderName,
			wantMessage: []string{"folder " + folderUUID(" ") + ": ", "folder " + folderUUID("  ") + ": "},
		},
		{
			name: "names every bad record at once",
			document: contents(
				[]record{folderRecord(" ", "")},
				[]record{tagRecord("a,b")},
				[]record{snippetRecord(record{"dateCreated": "yesterday"})},
			),
			wantErr: value.ErrBlankFolderName,
			wantMessage: []string{
				"folder " + folderUUID(" ") + ": ",
				"tag " + tagUUID("a,b") + ": ",
				"snippet " + snippetUUID + ": ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := newReader().Read(t.Context(), exportFile(t, tt.document))

			require.ErrorIs(t, err, tt.wantErr)
			require.ErrorContains(t, err, "snippetslab.Reader.Read: ")

			for _, want := range tt.wantMessage {
				assert.ErrorContains(t, err, want)
			}
		})
	}

	t.Run("refuses a dateDeleted that isn't a date and names the Snippet", func(t *testing.T) {
		t.Parallel()

		path := exportFile(t, contents(nil, nil, []record{snippetRecord(record{"dateDeleted": "last week"})}))

		_, _, err := newReader().Read(t.Context(), path)

		var parseErr *time.ParseError
		require.ErrorAs(t, err, &parseErr)
		assert.ErrorContains(t, err, "snippet "+snippetUUID+": parse deleted date: ")
	})

	t.Run("fails on a missing file", func(t *testing.T) {
		t.Parallel()

		_, _, err := newReader().Read(t.Context(), filepath.Join(t.TempDir(), exportName))

		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func tagRecord(name string) record {
	return record{"uuid": tagUUID(name), "title": name}
}

func tagUUID(name string) string {
	return "tag-" + name
}

func folderNamed(t *testing.T, library domain.Library, name string) domain.Folder {
	t.Helper()

	for _, folder := range library.Folders {
		if folder.Name().String() == name {
			return folder
		}
	}

	require.Failf(t, "no Folder named", "%q", name)

	return domain.Folder{}
}

func folderRecord(name, defaultLanguage string, children ...record) record {
	folder := record{"uuid": folderUUID(name), "title": name, "children": children}
	if defaultLanguage != "" {
		folder["defaultLanguage"] = defaultLanguage
	}

	return folder
}

func folderUUID(name string) string {
	return "folder-" + name
}

func fragmentTitle(t *testing.T, raw string) value.FragmentTitle {
	t.Helper()

	parsed, err := value.NewFragmentTitle(raw)
	require.NoError(t, err)

	return parsed
}

func title(t *testing.T, raw string) value.Title {
	t.Helper()

	parsed, err := value.NewTitle(raw)
	require.NoError(t, err)

	return parsed
}

func fragmentTitles(snippet domain.Snippet) []string {
	titles := make([]string, 0, len(snippet.Fragments()))
	for _, fragment := range snippet.Fragments() {
		titles = append(titles, fragment.Title().String())
	}

	return titles
}

func read(t *testing.T, path string) (domain.Library, domain.ImportLosses) {
	t.Helper()

	got, losses, err := newReader().Read(t.Context(), path)
	require.NoError(t, err)

	return got, losses
}

func newReader() *snippetslab.Reader {
	return snippetslab.NewReader(testkit.NewSequentialIDs(), testkit.NewFixedClock(importedAt()))
}

func importedAt() time.Time {
	return time.Date(2026, time.October, 14, 12, 0, 0, 0, time.UTC)
}

func exportFile(t *testing.T, document record) string {
	t.Helper()

	encoded, err := json.Marshal(document)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), exportName)
	require.NoError(t, os.WriteFile(path, encoded, 0o600))

	return path
}

func contents(folders, tags, snippets []record) record {
	return record{
		"app":    "SnippetsLab",
		"schema": "2.6",
		"contents": record{
			"folders":     folders,
			"tags":        tags,
			"snippets":    snippets,
			"smartGroups": []record{},
			"shortcuts":   []record{},
		},
	}
}

func snippetRecord(overrides record) record {
	return overridden(record{
		"uuid":         snippetUUID,
		"title":        "loop",
		"pinned":       false,
		"dateCreated":  created,
		"dateModified": modified,
		"fragments":    []record{fragmentRecord(nil)},
	}, overrides)
}

func fragmentRecord(overrides record) record {
	return overridden(record{
		"title":        "Fragment",
		"language":     "GoLexer",
		"content":      "for {}\n",
		"note":         "",
		"dateCreated":  created,
		"dateModified": created,
	}, overrides)
}

func attachmentRecord() record {
	return record{
		"uuid":       "A77AC000-0000-4000-8000-000000000001",
		"customName": "",
		"file":       record{"filename": "screenshot.png", "bytes": 2048, "checksum": "e3b0c442"},
	}
}

func overridden(defaults, overrides record) record {
	maps.Copy(defaults, overrides)

	return defaults
}
