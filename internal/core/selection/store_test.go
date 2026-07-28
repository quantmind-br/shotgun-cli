package selection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Load_MissingFile(t *testing.T) {
	t.Parallel()

	// Path inside a temp dir that was never written to.
	store := NewStore(filepath.Join(t.TempDir(), "selections.json"))

	deselected, err := store.Load("/some/project")

	require.NoError(t, err)
	assert.Nil(t, deselected)
}

func TestStore_SaveThenLoad_RoundTripsSorted(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")
	store := NewStore(path)
	project := "/home/user/proj"

	// Deliberately unsorted input; Save must persist ascending order.
	err := store.Save(project, []string{"c.go", "a.go", "b.go"})
	require.NoError(t, err)

	deselected, err := store.Load(project)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, deselected)

	// The backing file exists and decodes to the {"deselected":{...}} shape.
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var onDisk struct {
		Deselected map[string][]string `json:"deselected"`
	}
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, onDisk.Deselected[project])

	// A fresh Store over the same file reads back the same list.
	reopened := NewStore(path)
	deselected2, err := reopened.Load(project)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, deselected2)
}

func TestStore_Save_EmptyOrNilDeletesEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		empty []string
	}{
		{name: "empty slice", empty: []string{}},
		{name: "nil slice", empty: nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "selections.json")
			store := NewStore(path)
			project := "/home/user/proj"

			require.NoError(t, store.Save(project, []string{"a.go", "b.go"}))

			// Sanity: the entry is present before deletion.
			present, err := store.Load(project)
			require.NoError(t, err)
			require.Equal(t, []string{"a.go", "b.go"}, present)

			// Saving an empty/nil list deletes the project's entry.
			require.NoError(t, store.Save(project, tt.empty))

			deselected, err := store.Load(project)
			require.NoError(t, err)
			assert.Empty(t, deselected)

			// Deletion is reflected on disk, not just in memory.
			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			var onDisk struct {
				Deselected map[string][]string `json:"deselected"`
			}
			require.NoError(t, json.Unmarshal(raw, &onDisk))
			_, ok := onDisk.Deselected[project]
			assert.False(t, ok, "project entry should be removed from the file")
		})
	}
}

func TestStore_Save_ProjectsIsolatedInSharedFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")
	store := NewStore(path)
	projA := "/home/user/alpha"
	projB := "/home/user/beta"

	require.NoError(t, store.Save(projA, []string{"a1.go", "a2.go"}))
	require.NoError(t, store.Save(projB, []string{"b1.go"}))

	// Overwriting one project must not disturb the other.
	require.NoError(t, store.Save(projA, []string{"a3.go"}))

	gotA, err := store.Load(projA)
	require.NoError(t, err)
	assert.Equal(t, []string{"a3.go"}, gotA)

	gotB, err := store.Load(projB)
	require.NoError(t, err)
	assert.Equal(t, []string{"b1.go"}, gotB)

	// Deleting one project leaves the other intact.
	require.NoError(t, store.Save(projA, nil))

	gotA, err = store.Load(projA)
	require.NoError(t, err)
	assert.Empty(t, gotA)

	gotB, err = store.Load(projB)
	require.NoError(t, err)
	assert.Equal(t, []string{"b1.go"}, gotB)
}

func TestStore_Load_MalformedJSONReturnsError(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")
	require.NoError(t, os.WriteFile(path, []byte("{not valid json"), 0o600))

	store := NewStore(path)
	deselected, err := store.Load("/home/user/proj")

	require.Error(t, err)
	assert.Nil(t, deselected)
}

func TestStore_Save_DoesNotMutateCallerSlice(t *testing.T) {
	t.Parallel()

	store := NewStore(filepath.Join(t.TempDir(), "selections.json"))

	input := []string{"c.go", "a.go", "b.go"}
	require.NoError(t, store.Save("/home/user/proj", input))

	// The caller's slice ordering is untouched even though the store sorts.
	assert.Equal(t, []string{"c.go", "a.go", "b.go"}, input)
}

// Duas gravações concorrentes (dois projetos, mesmo store compartilhado em ~/.config)
// devem preservar as duas entradas.
func TestStore_ConcurrentSavesKeepEveryProject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sel.json")
	s := NewStore(path)

	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = s.Save(fmt.Sprintf("/proj/%d", i), []string{fmt.Sprintf("f%d.go", i)})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Logf("Save(%d) err = %v", i, err)
		}
	}
	missing := 0
	for i := range 8 {
		got, err := s.Load(fmt.Sprintf("/proj/%d", i))
		if err != nil {
			t.Fatalf("Load err: %v", err)
		}
		if len(got) == 0 {
			missing++
		}
	}
	if missing > 0 {
		t.Errorf("VIOLADO: %d/8 projetos perderam a seleção salva (race no read-modify-write + tmp fixo)", missing)
	}
}

// TestStore_Save_UnwritableDirReportsError: quando o diretório não aceita
// escrita, Save precisa falhar em vez de anunciar sucesso.
func TestStore_Save_UnwritableDirReportsError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignora permissões de diretório")
	}

	dir := t.TempDir()
	locked := filepath.Join(dir, "locked")
	require.NoError(t, os.Mkdir(locked, 0o500))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })

	s := NewStore(filepath.Join(locked, "sel.json"))
	err := s.Save("/proj", []string{"a.go"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write selection store")
}
