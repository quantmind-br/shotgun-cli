package selection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Store persists per-project file deselection preferences to a single JSON file.
type Store struct {
	path string
	// mu serializes the read-modify-write cycle of Save: the store file holds
	// every project, so concurrent saves would otherwise drop each other's entries.
	mu sync.Mutex
}

// NewStore returns a Store backed by the JSON file at path.
func NewStore(path string) *Store { return &Store{path: path} }

// storeData is the on-disk representation of all persisted project preferences.
type storeData struct {
	// Deselected maps a project's absolute path to its deselected relative paths (slash-form).
	Deselected map[string][]string `json:"deselected"`
}

// read loads and decodes the backing file. A missing or empty file yields an
// empty (non-nil) store; malformed JSON yields a wrapped error.
func (s *Store) read() (*storeData, error) {
	empty := &storeData{Deselected: map[string][]string{}}

	data, err := os.ReadFile(s.path)
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

// Load returns the deselected relative paths saved for projectPath.
// A missing file, empty file, or unknown project yields (nil, nil).
func (s *Store) Load(projectPath string) ([]string, error) {
	sd, err := s.read()
	if err != nil {
		return nil, err
	}
	return sd.Deselected[projectPath], nil
}

// Save replaces the deselected list for projectPath and persists the file.
// An empty/nil slice deletes the project's entry. The list is sorted before writing.
func (s *Store) Save(projectPath string, deselected []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, err := s.read()
	if err != nil {
		return err
	}

	if len(deselected) == 0 {
		delete(sd.Deselected, projectPath)
	} else {
		sorted := make([]string, len(deselected))
		copy(sorted, deselected)
		sort.Strings(sorted)
		sd.Deselected[projectPath] = sorted
	}

	out, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("encode selection store: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create selection store dir: %w", err)
	}

	// A unique temp file: a fixed ".tmp" name lets a second writer delete the
	// file this one is about to rename.
	tmp, err := os.CreateTemp(dir, filepath.Base(s.path)+".tmp*")
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
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("commit selection store: %w", err)
	}

	return nil
}
