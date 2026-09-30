// PROTOTYPE — throwaway. In-memory fake data; nothing is persisted.
package main

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

type folder struct {
	id, name, parent, lang string
	collapsed              bool
}

type snippet struct {
	id, title, desc, lang, content, folder string
	tags                                   []string
	created, updated                       time.Time
}

type store struct {
	folders  []*folder
	snippets []*snippet
	tags     []string
	nextID   int
}

func (s *store) newID() string {
	s.nextID++
	return fmt.Sprintf("id%d", s.nextID)
}

func seed(empty bool) *store {
	s := &store{}
	if empty {
		return s
	}
	day := func(n int) time.Time { return time.Date(2026, 9, n, 10, 0, 0, 0, time.Local) }
	s.folders = []*folder{
		{id: "docker", name: "docker", lang: "Docker"},
		{id: "compose", name: "compose", parent: "docker", lang: "YAML"},
		{id: "go", name: "go", lang: "Go"},
		{id: "gotest", name: "testing", parent: "go", lang: "Go"},
		{id: "shell", name: "shell", lang: "Bash"},
		{id: "empty", name: "empty folder", lang: "plaintext"},
	}
	s.tags = []string{"docker", "git", "go", "oneliner", "testing", "unused"}
	add := func(folderID, title, desc, lang string, tags []string, created, updated int, content string) {
		s.snippets = append(s.snippets, &snippet{
			id: s.newID(), folder: folderID, title: title, desc: desc, lang: lang, tags: tags,
			created: day(created), updated: day(updated), content: strings.TrimPrefix(content, "\n"),
		})
	}
	add("docker", "Multi-stage Go build", "Small final image for a Go binary", "Docker", []string{"docker", "go"}, 2, 20, `
FROM golang:1.27 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/app

FROM gcr.io/distroless/static
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
`)
	add("docker", "Prune everything", "Reclaim disk space", "Bash", []string{"docker", "oneliner"}, 3, 3, `
docker system prune --all --volumes --force
`)
	add("compose", "Postgres for local dev", "", "YAML", []string{"docker"}, 5, 18, `
services:
  db:
    image: postgres:17
    environment:
      POSTGRES_PASSWORD: dev
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
volumes:
  pgdata:
`)
	add("go", "Graceful HTTP shutdown", "Stop accepting, drain, exit", "Go", []string{"go"}, 6, 27, `
func run(ctx context.Context, srv *http.Server) error {
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
`)
	add("go", "errors.Join a slice", "", "Go", []string{"go"}, 7, 7, `
var errs []error
for _, f := range fields {
	if err := f.Validate(); err != nil {
		errs = append(errs, err)
	}
}
return errors.Join(errs...)
`)
	add("gotest", "Table test skeleton", "Parallel table-driven test", "Go", []string{"go", "testing"}, 8, 29, `
func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty", input: "", want: 0},
		{name: "one", input: "1", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Parse(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
`)
	add("shell", "Undo last commit, keep changes", "", "Bash", []string{"git", "oneliner"}, 9, 9, `
git reset --soft HEAD~1
`)
	add("shell", "Find large files", "Top 20 files by size under the current directory", "Bash", []string{"oneliner"}, 10, 21, `
find . -type f -printf '%s %p\n' | sort -nr | head -20 | numfmt --field=1 --to=iec
`)
	add("", "SSH config host block", "Jump host through a bastion", "SSH Config", nil, 11, 11, `
Host internal
    HostName 10.0.4.12
    User deploy
    ProxyJump bastion.example.com
`)
	add("", "Meeting notes template", "", "plaintext", nil, 12, 12, `
Attendees:
Decisions:
Action items:
`)
	return s
}

func (s *store) folder(id string) *folder {
	for _, f := range s.folders {
		if f.id == id {
			return f
		}
	}
	return nil
}

func (s *store) children(parent string) []*folder {
	var out []*folder
	for _, f := range s.folders {
		if f.parent == parent {
			out = append(out, f)
		}
	}
	slices.SortFunc(out, func(a, b *folder) int { return strings.Compare(a.name, b.name) })
	return out
}

func (s *store) path(folderID string) string {
	if folderID == "" {
		return "Root"
	}
	var parts []string
	for f := s.folder(folderID); f != nil; f = s.folder(f.parent) {
		parts = append([]string{f.name}, parts...)
	}
	return "Root / " + strings.Join(parts, " / ")
}

func (s *store) subtree(id string) []string {
	ids := []string{id}
	for _, c := range s.children(id) {
		ids = append(ids, s.subtree(c.id)...)
	}
	return ids
}

func (s *store) isDescendant(candidate, of string) bool {
	return slices.Contains(s.subtree(of), candidate)
}

func (s *store) inFolder(id string) []*snippet {
	var out []*snippet
	for _, sn := range s.snippets {
		if sn.folder == id {
			out = append(out, sn)
		}
	}
	return out
}

func (s *store) withTag(tag string) []*snippet {
	var out []*snippet
	for _, sn := range s.snippets {
		if slices.Contains(sn.tags, tag) {
			out = append(out, sn)
		}
	}
	return out
}

func (s *store) folderCount(id string) int { return len(s.inFolder(id)) }

func (s *store) tagCount(tag string) int { return len(s.withTag(tag)) }

func (s *store) deleteFolder(id string) (subfolders, snippets int) {
	ids := s.subtree(id)
	s.folders = slices.DeleteFunc(s.folders, func(f *folder) bool { return slices.Contains(ids, f.id) })
	before := len(s.snippets)
	s.snippets = slices.DeleteFunc(s.snippets, func(sn *snippet) bool { return slices.Contains(ids, sn.folder) })
	return len(ids) - 1, before - len(s.snippets)
}

func (s *store) subtreeCounts(id string) (subfolders, snippets int) {
	ids := s.subtree(id)
	for _, fid := range ids {
		snippets += s.folderCount(fid)
	}
	return len(ids) - 1, snippets
}

func (s *store) deleteSnippet(id string) {
	s.snippets = slices.DeleteFunc(s.snippets, func(sn *snippet) bool { return sn.id == id })
}

func (s *store) deleteTag(tag string) {
	s.tags = slices.DeleteFunc(s.tags, func(t string) bool { return t == tag })
	for _, sn := range s.snippets {
		sn.tags = slices.DeleteFunc(sn.tags, func(t string) bool { return t == tag })
	}
}

func (s *store) ensureTags(tags []string) {
	for _, t := range tags {
		if !slices.ContainsFunc(s.tags, func(x string) bool { return strings.EqualFold(x, t) }) {
			s.tags = append(s.tags, t)
		}
	}
	slices.SortFunc(s.tags, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
}

// search fakes the spec's weighting (title ×4, Tag ×3, Description ×2, content ×1)
// with subsequence matching; the real scorer is out of scope here.
func (s *store) search(query string) []*snippet {
	type scored struct {
		sn    *snippet
		score int
	}
	var hits []scored
	for _, sn := range s.snippets {
		best := 0
		consider := func(weight int, text string, fuzzy bool) {
			if matches(text, query, fuzzy) && weight > best {
				best = weight
			}
		}
		consider(4, sn.title, true)
		for _, t := range sn.tags {
			consider(3, t, true)
		}
		consider(2, sn.desc, true)
		consider(1, sn.content, false)
		if best > 0 {
			hits = append(hits, scored{sn, best})
		}
	}
	slices.SortStableFunc(hits, func(a, b scored) int {
		if a.score != b.score {
			return b.score - a.score
		}
		return b.sn.updated.Compare(a.sn.updated)
	})
	out := make([]*snippet, len(hits))
	for i, h := range hits {
		out[i] = h.sn
	}
	return out
}

func matches(text, query string, fuzzy bool) bool {
	text, query = strings.ToLower(text), strings.ToLower(query)
	if !fuzzy {
		return strings.Contains(text, query)
	}
	i := 0
	for _, r := range text {
		if i < len(query) && r == rune(query[i]) {
			i++
		}
	}
	return i == len(query)
}

var languages = []string{
	"plaintext", "Bash", "Docker", "Go", "JSON", "Makefile", "Markdown", "PHP",
	"Python", "Rust", "SQL", "SSH Config", "TOML", "TypeScript", "YAML",
}
