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

	// The project's own file exists and records which project it belongs to.
	raw, err := os.ReadFile(store.projectFile(project))
	require.NoError(t, err)

	var onDisk struct {
		Project    string   `json:"project"`
		Deselected []string `json:"deselected"`
	}
	require.NoError(t, json.Unmarshal(raw, &onDisk))
	assert.Equal(t, project, onDisk.Project)
	assert.Equal(t, []string{"a.go", "b.go", "c.go"}, onDisk.Deselected)

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
			_, err = os.Stat(store.projectFile(project))
			assert.True(t, os.IsNotExist(err), "the project's file should be removed")
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
	// The failure now surfaces when creating the per-project directory, which is
	// the first write the store attempts.
	assert.Contains(t, err.Error(), "selection store")
}

// TestStore_ConcurrentProcessesDoNotDropEachOther cobre a suspeita que o bug hunt
// anterior deixou aberta: o mutex de Store só ordena escritores dentro de um
// binário, e o arquivo guarda todos os projetos. Duas instâncias de shotgun-cli
// em projetos diferentes liam o mesmo conteúdo e o segundo `rename` descartava a
// entrada do primeiro.
//
// Dois *Store distintos sobre o mesmo caminho é exatamente o que duas instâncias
// veem: não compartilham mutex, só o arquivo.
func TestStore_ConcurrentProcessesDoNotDropEachOther(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")

	const projects = 8
	var wg sync.WaitGroup
	errs := make(chan error, projects)

	for i := 0; i < projects; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// Um Store por goroutine: sem mutex compartilhado, como instâncias distintas.
			s := NewStore(path)
			if err := s.Save(fmt.Sprintf("/proj/%d", n), []string{fmt.Sprintf("f%d.go", n)}); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("Save falhou sob concorrência: %v", err)
	}

	reader := NewStore(path)
	for i := 0; i < projects; i++ {
		got, err := reader.Load(fmt.Sprintf("/proj/%d", i))
		require.NoError(t, err)
		assert.Equal(t, []string{fmt.Sprintf("f%d.go", i)}, got,
			"a entrada do projeto %d foi perdida por outro escritor", i)
	}
}

// TestStore_ConcurrentSameProjectConverges garante que a disputa pelo mesmo
// projeto termina num dos valores gravados, e não num arquivo corrompido ou
// vazio.
func TestStore_ConcurrentSameProjectConverges(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = NewStore(path).Save("/proj", []string{fmt.Sprintf("f%d.go", n)})
		}(i)
	}
	wg.Wait()

	got, err := NewStore(path).Load("/proj")
	require.NoError(t, err)
	require.Len(t, got, 1, "o arquivo deve conter exatamente um dos valores gravados")
	assert.Regexp(t, `^f[0-5]\.go$`, got[0])
}

// TestStore_MigratesLegacySingleFile cobre o caminho que toca dados de usuários
// que já usavam a versão anterior: o arquivo único é dividido em arquivos por
// projeto e removido, sem perder nada.
func TestStore_MigratesLegacySingleFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "selections.json")
	legacy := `{"deselected":{"/proj/a":["a1.go","a2.go"],"/proj/b":["b1.go"]}}`
	require.NoError(t, os.WriteFile(path, []byte(legacy), 0o600))

	store := NewStore(path)

	gotA, err := store.Load("/proj/a")
	require.NoError(t, err)
	assert.Equal(t, []string{"a1.go", "a2.go"}, gotA)

	gotB, err := store.Load("/proj/b")
	require.NoError(t, err)
	assert.Equal(t, []string{"b1.go"}, gotB)

	// O arquivo legado sai de cena depois de migrado.
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err), "o arquivo legado deve ser removido após a migração")

	// E os dados sobrevivem a uma reabertura, agora vindos dos arquivos novos.
	reopened := NewStore(path)
	gotA2, err := reopened.Load("/proj/a")
	require.NoError(t, err)
	assert.Equal(t, []string{"a1.go", "a2.go"}, gotA2)
}

// TestStore_MigrationDoesNotClobberNewerData garante que a migração não sobrescreve
// um arquivo por projeto já gravado pelo código novo — ele é mais recente que
// qualquer coisa no arquivo legado.
func TestStore_MigrationDoesNotClobberNewerData(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "selections.json")

	store := NewStore(path)
	require.NoError(t, store.Save("/proj/a", []string{"novo.go"}))

	// Só agora aparece um arquivo legado com um valor antigo para o mesmo projeto.
	require.NoError(t, os.WriteFile(path, []byte(`{"deselected":{"/proj/a":["antigo.go"]}}`), 0o600))

	got, err := NewStore(path).Load("/proj/a")
	require.NoError(t, err)
	assert.Equal(t, []string{"novo.go"}, got, "o dado novo não pode ser sobrescrito pelo legado")
}

// TestStore_MigrationReportsCorruptLegacyFile mantém o comportamento anterior:
// um arquivo legado corrompido é reportado, não descartado em silêncio.
func TestStore_MigrationReportsCorruptLegacyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "selections.json")
	require.NoError(t, os.WriteFile(path, []byte("{nao é json"), 0o600))

	_, err := NewStore(path).Load("/proj")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse selection store")
}
