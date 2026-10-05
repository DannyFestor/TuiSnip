package domain_test

import (
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
	"github.com/DannyFestor/TuiSnip/internal/testkit"
)

func TestFolder_New(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        domain.FolderID
		parentID  domain.FolderID
		updatedAt time.Time
		wantErrs  []error
	}{
		{name: "accepts a Folder at the Root", id: folderID(), parentID: domain.FolderID{}, updatedAt: created},
		{name: "accepts a Folder in another", id: folderID(), parentID: parentID(), updatedAt: created},
		{
			name:      "rejects the nil id",
			id:        domain.FolderID{},
			parentID:  parentID(),
			updatedAt: created,
			wantErrs:  []error{domain.ErrNilID},
		},
		{
			name:      "rejects being its own parent",
			id:        folderID(),
			parentID:  folderID(),
			updatedAt: created,
			wantErrs:  []error{domain.ErrFolderCycle},
		},
		{
			name:      "reports every broken rule",
			id:        folderID(),
			parentID:  folderID(),
			updatedAt: created.Add(-time.Second),
			wantErrs:  []error{domain.ErrFolderCycle, domain.ErrUpdatedBeforeCreate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewFolder(
				tt.id, mustFolderName(t, "docker"), tt.parentID, value.PlainText(), created, tt.updatedAt,
			)

			requireErrors(t, err, tt.wantErrs)
		})
	}
}

func TestFolder_Accessors(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	name := mustFolderName(t, "docker")
	language := mustLanguage(t, "Bash")

	folder, err := domain.NewFolder(folderID(), name, parentID(), language, created, updated)

	require.NoError(t, err)
	assert.Equal(t, folderID(), folder.ID())
	assert.Equal(t, name, folder.Name())
	assert.Equal(t, parentID(), folder.ParentID())
	assert.Equal(t, language, folder.DefaultLanguage())
	assert.Equal(t, created, folder.CreatedAt())
	assert.Equal(t, updated, folder.UpdatedAt())
}

func TestFolder_AtRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parentID domain.FolderID
		want     bool
	}{
		{name: "zero parent id is the Root", parentID: domain.FolderID{}, want: true},
		{name: "a parent id is not the Root", parentID: parentID(), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			folder := testkit.Folder(t, testkit.FolderSpec{ParentID: tt.parentID})

			assert.Equal(t, tt.want, folder.AtRoot())
		})
	}
}

func TestFolder_NewAtRoot(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)

	folder, err := domain.NewFolderAtRoot(folderID(), mustFolderName(t, "docker"), now)

	require.NoError(t, err)
	assert.True(t, folder.AtRoot())
	assert.Equal(t, value.PlainText(), folder.DefaultLanguage())
	assert.Equal(t, now, folder.CreatedAt())
	assert.Equal(t, now, folder.UpdatedAt())
}

func TestFolder_NewIn(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	parent := testkit.Folder(t, testkit.FolderSpec{ID: parentID(), DefaultLanguage: "Go"})

	folder, err := domain.NewFolderIn(parent, folderID(), mustFolderName(t, "generics"), now)

	require.NoError(t, err)
	assert.Equal(t, parentID(), folder.ParentID())
	assert.Equal(t, "Go", folder.DefaultLanguage().String())
	assert.Equal(t, now, folder.CreatedAt())
}

func TestFolder_MoveUnder(t *testing.T) {
	t.Parallel()

	descendantID := domain.FolderID(uuid.MustParse("0192f0c1-7a3b-7c4d-8e5f-00000000d35c"))

	tests := []struct {
		name         string
		parentID     domain.FolderID
		wantParentID domain.FolderID
		wantErrs     []error
	}{
		{name: "moves under an unrelated Folder", parentID: parentID(), wantParentID: parentID()},
		{name: "rejects moving under itself", parentID: folderID(), wantErrs: []error{domain.ErrFolderCycle}},
		{
			name:     "rejects moving under a descendant",
			parentID: descendantID,
			wantErrs: []error{domain.ErrFolderCycle},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			folder := testkit.Folder(t, testkit.FolderSpec{ID: folderID()})

			moved, err := folder.MoveUnder(tt.parentID, []domain.FolderID{descendantID})

			requireErrors(t, err, tt.wantErrs)
			assert.Equal(t, tt.wantParentID, moved.ParentID())
			assert.True(t, folder.AtRoot(), "the original is untouched")
		})
	}
}

func TestFolder_MoveUnderKeepsUpdatedAt(t *testing.T) {
	t.Parallel()

	updated := time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC)
	folder := testkit.Folder(t, testkit.FolderSpec{ID: folderID(), UpdatedAt: updated})

	moved, err := folder.MoveUnder(parentID(), nil)

	require.NoError(t, err)
	assert.Equal(t, updated, moved.UpdatedAt())
}

func TestFolder_MoveToRoot(t *testing.T) {
	t.Parallel()

	folder := testkit.Folder(t, testkit.FolderSpec{ID: folderID(), ParentID: parentID()})

	moved := folder.MoveToRoot()

	assert.True(t, moved.AtRoot())
	assert.False(t, folder.AtRoot(), "the original is untouched")
	assert.Equal(t, folder.UpdatedAt(), moved.UpdatedAt())
}

func TestFolder_Rename(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC)
	folder := testkit.Folder(t, testkit.FolderSpec{ID: folderID(), Name: "go", CreatedAt: created})

	t.Run("takes the name and the time", func(t *testing.T) {
		t.Parallel()

		renamed, err := folder.Rename(mustFolderName(t, "golang"), created.Add(time.Hour))

		require.NoError(t, err)
		assert.Equal(t, "golang", renamed.Name().String())
		assert.Equal(t, created.Add(time.Hour), renamed.UpdatedAt())
		assert.Equal(t, "go", folder.Name().String(), "the original is untouched")
	})

	t.Run("rejects a time before creation", func(t *testing.T) {
		t.Parallel()

		_, err := folder.Rename(mustFolderName(t, "golang"), created.Add(-time.Hour))

		require.ErrorIs(t, err, domain.ErrUpdatedBeforeCreate)
	})
}

func TestCompareFolders(t *testing.T) {
	t.Parallel()

	ids := testkit.NewSequentialIDs()
	first, second := ids.NewFolderID(), ids.NewFolderID()

	tests := []struct {
		name      string
		names     [2]string
		wantOrder []domain.FolderID
	}{
		{name: "orders by name", names: [2]string{"zsh", "awk"}, wantOrder: []domain.FolderID{second, first}},
		{name: "ignores case", names: [2]string{"Zsh", "awk"}, wantOrder: []domain.FolderID{second, first}},
		{name: "breaks a name tie by ID", names: [2]string{"go", "Go"}, wantOrder: []domain.FolderID{first, second}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			folders := []domain.Folder{
				testkit.Folder(t, testkit.FolderSpec{ID: second, Name: tt.names[1]}),
				testkit.Folder(t, testkit.FolderSpec{ID: first, Name: tt.names[0]}),
			}

			slices.SortFunc(folders, domain.CompareFolders)

			assert.Equal(t, tt.wantOrder, []domain.FolderID{folders[0].ID(), folders[1].ID()})
		})
	}
}

func TestFolderTreeMoveProperty(t *testing.T) {
	t.Parallel()

	// Parsed once here, because *rapid.T is no testing.TB and can't run the must helpers.
	name := mustFolderName(t, "folder")
	languages := []value.Language{
		value.PlainText(), mustLanguage(t, "Go"), mustLanguage(t, "Bash"), mustLanguage(t, "Python"),
	}

	rapid.Check(t, func(rt *rapid.T) {
		rt.Repeat(newFolderTreeModel(name, languages).actions())
	})
}

type folderTreeModel struct {
	ids                *testkit.SequentialIDs
	now                time.Time
	name               value.FolderName
	languages          []value.Language
	order              []domain.FolderID
	folders            map[domain.FolderID]domain.Folder
	languageAtCreation map[domain.FolderID]value.Language
}

func newFolderTreeModel(name value.FolderName, languages []value.Language) *folderTreeModel {
	return &folderTreeModel{
		ids:                testkit.NewSequentialIDs(),
		now:                time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC),
		name:               name,
		languages:          languages,
		order:              nil,
		folders:            map[domain.FolderID]domain.Folder{},
		languageAtCreation: map[domain.FolderID]value.Language{},
	}
}

func (m *folderTreeModel) actions() map[string]func(*rapid.T) {
	return map[string]func(*rapid.T){
		"create at the Root":                 m.createAtRoot,
		"create at the Root with a Language": m.createAtRootWithLanguage,
		"create in a Folder":                 m.createIn,
		"move under a Folder":                m.moveUnder,
		"move to the Root":                   m.moveToRoot,
		"":                                   m.checkInvariants,
	}
}

func (m *folderTreeModel) createAtRoot(rt *rapid.T) {
	folder, err := domain.NewFolderAtRoot(m.ids.NewFolderID(), m.name, m.now)
	require.NoError(rt, err)

	m.add(folder)
}

func (m *folderTreeModel) createAtRootWithLanguage(rt *rapid.T) {
	language := rapid.SampledFrom(m.languages).Draw(rt, "language")
	folder, err := domain.NewFolder(m.ids.NewFolderID(), m.name, domain.FolderID{}, language, m.now, m.now)
	require.NoError(rt, err)

	m.add(folder)
}

func (m *folderTreeModel) createIn(rt *rapid.T) {
	parent := m.folders[m.drawID(rt, "parent")]

	folder, err := domain.NewFolderIn(parent, m.ids.NewFolderID(), m.name, m.now)
	require.NoError(rt, err)
	require.Equal(rt, parent.DefaultLanguage(), folder.DefaultLanguage())

	m.add(folder)
}

func (m *folderTreeModel) moveUnder(rt *rapid.T) {
	folder := m.folders[m.drawID(rt, "folder")]
	parentID := m.drawID(rt, "parent")
	descendantIDs := m.descendantsOf(folder.ID())

	moved, err := folder.MoveUnder(parentID, descendantIDs)

	if parentID == folder.ID() || m.hasAncestor(parentID, folder.ID()) {
		require.ErrorIs(rt, err, domain.ErrFolderCycle)

		return
	}

	require.NoError(rt, err)

	m.folders[moved.ID()] = moved
}

func (m *folderTreeModel) moveToRoot(rt *rapid.T) {
	folder := m.folders[m.drawID(rt, "folder")]

	m.folders[folder.ID()] = folder.MoveToRoot()
}

func (m *folderTreeModel) checkInvariants(rt *rapid.T) {
	for _, id := range m.order {
		folder := m.folders[id]
		require.True(rt, m.reachesRoot(folder), "Folder %s is cut off from the Root", id)
		require.Equal(rt, m.languageAtCreation[id], folder.DefaultLanguage())
	}
}

func (m *folderTreeModel) add(folder domain.Folder) {
	m.order = append(m.order, folder.ID())
	m.folders[folder.ID()] = folder
	m.languageAtCreation[folder.ID()] = folder.DefaultLanguage()
}

func (m *folderTreeModel) drawID(rt *rapid.T, label string) domain.FolderID {
	if len(m.order) == 0 {
		rt.Skip("no Folder yet")
	}

	return rapid.SampledFrom(m.order).Draw(rt, label)
}

func (m *folderTreeModel) descendantsOf(ancestorID domain.FolderID) []domain.FolderID {
	var descendantIDs []domain.FolderID

	for _, id := range m.order {
		if id != ancestorID && m.hasAncestor(id, ancestorID) {
			descendantIDs = append(descendantIDs, id)
		}
	}

	return descendantIDs
}

// Walking is capped at the Folder count, so a cycle the domain let through ends the
// walk instead of looping forever, and reachesRoot then reports it.
func (m *folderTreeModel) hasAncestor(id, ancestorID domain.FolderID) bool {
	current := m.folders[id]
	for range len(m.order) {
		if current.AtRoot() {
			return false
		}

		if current.ParentID() == ancestorID {
			return true
		}

		current = m.folders[current.ParentID()]
	}

	return false
}

func (m *folderTreeModel) reachesRoot(folder domain.Folder) bool {
	current := folder
	for range len(m.order) {
		if current.AtRoot() {
			return true
		}

		current = m.folders[current.ParentID()]
	}

	return false
}

func folderID() domain.FolderID {
	return domain.FolderID(uuid.MustParse(storedID))
}

func parentID() domain.FolderID {
	return domain.FolderID(uuid.MustParse(folderUUID))
}

func mustFolderName(tb testing.TB, raw string) value.FolderName {
	tb.Helper()

	name, err := value.NewFolderName(raw)
	require.NoError(tb, err)

	return name
}

func mustLanguage(tb testing.TB, raw string) value.Language {
	tb.Helper()

	language, err := value.NewLanguage(raw)
	require.NoError(tb, err)

	return language
}
