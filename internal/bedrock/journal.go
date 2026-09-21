package bedrock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const journalVersion = 1

type JournalState string

const (
	JournalPrepared  JournalState = "PREPARED"
	JournalMutating  JournalState = "MUTATING"
	JournalCompleted JournalState = "COMPLETED"
)

type JournalOriginal struct {
	Path string `json:"path"`
	Existed bool `json:"existed"`
	Mode uint32 `json:"mode,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
	Content []byte `json:"content,omitempty"`
}

type RunJournal struct {
	Version int `json:"version"`
	RunID string `json:"run_id"`
	Repository string `json:"repository"`
	State JournalState `json:"state"`
	Originals []JournalOriginal `json:"originals"`
	RecordedAtUTC string `json:"recorded_at_utc"`
}

func PrepareRunJournal(root, runID string, paths []string) (RunJournal, string, error) {
	if err := validateRunID(runID); err != nil { return RunJournal{}, "", err }
	root, err := filepath.Abs(root); if err != nil { return RunJournal{}, "", err }
	journal := RunJournal{Version: journalVersion, RunID: runID, Repository: root, State: JournalPrepared, RecordedAtUTC: time.Now().UTC().Format(time.RFC3339Nano)}
	seen := map[string]struct{}{}
	for _, requested := range paths {
		rel, target, err := secureTarget(root, requested); if err != nil { return RunJournal{}, "", err }
		if _, duplicate := seen[rel]; duplicate { return RunJournal{}, "", fmt.Errorf("duplicate journal path %q", rel) }
		seen[rel] = struct{}{}
		original := JournalOriginal{Path: rel}
		data, err := os.ReadFile(target)
		if errors.Is(err, os.ErrNotExist) { journal.Originals = append(journal.Originals, original); continue }
		if err != nil { return RunJournal{}, "", fmt.Errorf("read original %q: %w", rel, err) }
		info, err := os.Stat(target); if err != nil { return RunJournal{}, "", fmt.Errorf("stat original %q: %w", rel, err) }
		sum := sha256.Sum256(data)
		original.Existed, original.Mode, original.SHA256, original.Content = true, uint32(info.Mode().Perm()), hex.EncodeToString(sum[:]), data
		journal.Originals = append(journal.Originals, original)
	}
	path, err := saveRunJournal(journal); if err != nil { return RunJournal{}, "", err }
	return journal, path, nil
}

// TransitionRunJournal durably advances a prepared journal through the mutation
// boundary. Only forward transitions are accepted so a completed run cannot be
// made to look interrupted by a later caller.
func TransitionRunJournal(path string, next JournalState) (RunJournal, error) {
	data, err := os.ReadFile(path)
	if err != nil { return RunJournal{}, fmt.Errorf("read run journal: %w", err) }
	var journal RunJournal
	if err := json.Unmarshal(data, &journal); err != nil { return RunJournal{}, fmt.Errorf("decode run journal: %w", err) }
	if journal.Version != journalVersion { return RunJournal{}, fmt.Errorf("unsupported journal version %d", journal.Version) }
	if err := validateRunID(journal.RunID); err != nil { return RunJournal{}, err }
	allowed := (journal.State == JournalPrepared && next == JournalMutating) || (journal.State == JournalMutating && next == JournalCompleted)
	if !allowed { return RunJournal{}, fmt.Errorf("invalid journal transition %s -> %s", journal.State, next) }
	journal.State = next
	if err := persistRunJournalAt(path, journal); err != nil { return RunJournal{}, err }
	return journal, nil
}

func validateRunID(runID string) error {
	if runID == "" { return errors.New("run id is required") }
	if runID == "." || runID == ".." || filepath.Base(runID) != runID || strings.ContainsAny(runID, `/\\`) { return fmt.Errorf("unsafe run id %q", runID) }
	return nil
}

func saveRunJournal(journal RunJournal) (string, error) {
	if err := validateRunID(journal.RunID); err != nil { return "", err }
	cacheDir, err := os.UserCacheDir(); if err != nil { return "", err }
	repoSum := sha256.Sum256([]byte(journal.Repository))
	dir := filepath.Join(cacheDir, "bedrock", "journals", hex.EncodeToString(repoSum[:8]))
	if err := os.MkdirAll(dir, 0o700); err != nil { return "", err }
	final := filepath.Join(dir, journal.RunID+".json")
	if err := persistRunJournalAt(final, journal); err != nil { return "", err }
	return final, nil
}

func persistRunJournalAt(final string, journal RunJournal) error {
	data, err := json.MarshalIndent(journal, "", "  "); if err != nil { return err }
	dir := filepath.Dir(final)
	tmp, err := os.CreateTemp(dir, journal.RunID+"-*.tmp"); if err != nil { return err }
	tmpName := tmp.Name(); defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil { tmp.Close(); return err }
	if _, err := tmp.Write(data); err != nil { tmp.Close(); return err }
	if err := tmp.Sync(); err != nil { tmp.Close(); return err }
	if err := tmp.Close(); err != nil { return err }
	if err := os.Rename(tmpName, final); err != nil { return err }
	return nil
}
