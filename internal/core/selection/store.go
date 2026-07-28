package selection

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Store persists per-project file deselection preferences.
//
// Each project gets its own file. A single shared file was the earlier design,
// and it could not be made safe across processes: the store holds every project,
// so two shotgun-cli instances working on different projects would each
// read-modify-write the whole map and the second rename would drop the first's
// entry. An in-process mutex cannot see the other process, and an optimistic
// re-read still leaves a window between the check and the rename -- measurably,
// under eight concurrent writers.
//
// Separate files remove the interference by construction rather than narrowing
// it: writers on different projects never touch the same path. Two writers on the
// *same* project still race, and the last atomic rename wins, which is the
// outcome a user would expect anyway.
type Store struct {
	// legacyPath is the pre-existing single-file store, read once and migrated.
	legacyPath string
	// dir holds one file per project.
	dir string
	// mu serializes this process's writers, including the one-shot migration.
	mu sync.Mutex
}

// NewStore returns a Store. path is the legacy single-file location; per-project
// files live in a "selections" directory beside it.
func NewStore(path string) *Store {
	return &Store{
		legacyPath: path,
		dir:        filepath.Join(filepath.Dir(path), "selections"),
	}
}

// storeData is the on-disk representation of the legacy single-file store.
type storeData struct {
	// Deselected maps a project's absolute path to its deselected relative paths (slash-form).
	Deselected map[string][]string `json:"deselected"`
}

// projectData is the on-disk representation of one project's preferences.
type projectData struct {
	// Project records which path this file belongs to, so the directory stays
	// readable by a human despite the hashed filenames.
	Project    string   `json:"project"`
	Deselected []string `json:"deselected"`
}

// projectFile maps a project path to its own file. The hash keeps the name
// filesystem-safe and length-bounded on every platform.
func (s *Store) projectFile(projectPath string) string {
	sum := sha256.Sum256([]byte(projectPath))

	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+".json")
}

// readLegacy loads the legacy single-file store. A missing or empty file yields
// an empty (non-nil) store; malformed JSON yields a wrapped error.
func (s *Store) readLegacy() (*storeData, error) {
	empty := &storeData{Deselected: map[string][]string{}}

	data, err := os.ReadFile(s.legacyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return empty, nil
		}

		return nil, fmt.Errorf("read selection store: %w", err)
	}
	if len(data) == 0 {
		return empty, nil
	}

	var sd storeData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("parse selection store: %w", err)
	}
	if sd.Deselected == nil {
		sd.Deselected = map[string][]string{}
	}

	return &sd, nil
}

// migrateLegacy splits the legacy file into per-project files and removes it.
//
// Idempotent, and safe if two processes attempt it at once: both write identical
// content from the same source, and only one of them succeeds at removing the
// legacy file. Caller must hold mu.
func (s *Store) migrateLegacy() error {
	// A corrupt legacy file is reported rather than silently discarded: it holds
	// the user's saved deselections, and the previous store surfaced the same
	// error. They can delete the file to move on.
	sd, err := s.readLegacy()
	if err != nil {
		return err
	}
	if len(sd.Deselected) == 0 {
		// Nothing to carry over. Remove an empty legacy file so this is a no-op
		// on every later call.
		_ = os.Remove(s.legacyPath)

		return nil
	}

	for project, deselected := range sd.Deselected {
		// Do not clobber a per-project file already written by the new code: it
		// is newer than anything the legacy file holds.
		if _, err := os.Stat(s.projectFile(project)); err == nil {
			continue
		}
		if err := s.writeProject(project, deselected); err != nil {
			return err
		}
	}

	_ = os.Remove(s.legacyPath)

	return nil
}

// Load returns the deselected relative paths saved for projectPath.
// A missing file, empty file, or unknown project yields (nil, nil).
func (s *Store) Load(projectPath string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.migrateLegacy(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.projectFile(projectPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("read selection store: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}

	var pd projectData
	if err := json.Unmarshal(data, &pd); err != nil {
		return nil, fmt.Errorf("parse selection store: %w", err)
	}
	if len(pd.Deselected) == 0 {
		return nil, nil
	}

	return pd.Deselected, nil
}

// Save replaces the deselected list for projectPath and persists it.
// An empty/nil slice deletes the project's entry. The list is sorted before writing.
func (s *Store) Save(projectPath string, deselected []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.migrateLegacy(); err != nil {
		return err
	}

	if len(deselected) == 0 {
		if err := os.Remove(s.projectFile(projectPath)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("commit selection store: %w", err)
		}

		return nil
	}

	return s.writeProject(projectPath, deselected)
}

// writeProject atomically replaces one project's file. Caller must hold mu.
func (s *Store) writeProject(projectPath string, deselected []string) error {
	sorted := make([]string, len(deselected))
	copy(sorted, deselected)
	sort.Strings(sorted)

	out, err := json.MarshalIndent(projectData{Project: projectPath, Deselected: sorted}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode selection store: %w", err)
	}

	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return fmt.Errorf("create selection store dir: %w", err)
	}

	// A unique temp file: a fixed ".tmp" name lets a second writer delete the
	// file this one is about to rename.
	tmp, err := os.CreateTemp(s.dir, "store.tmp*")
	if err != nil {
		return fmt.Errorf("write selection store: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once the rename succeeded

	if _, err := tmp.Write(out); err != nil {
		_ = tmp.Close()

		return fmt.Errorf("write selection store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write selection store: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fmt.Errorf("write selection store: %w", err)
	}
	if err := os.Rename(tmpName, s.projectFile(projectPath)); err != nil {
		return fmt.Errorf("commit selection store: %w", err)
	}

	return nil
}
