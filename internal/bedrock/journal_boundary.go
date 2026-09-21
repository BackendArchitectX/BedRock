package bedrock

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// recordCapturedMutationIntent persists mutation ownership from an original
// captured by the write boundary. It deliberately does not re-read the target:
// the caller must use the same capture for its in-memory rollback state, so
// durable recovery and ordinary rollback cannot disagree about pre-write bytes.
func recordCapturedMutationIntent(path string, original JournalOriginal, intended []byte) (RunJournal, error) {
	journal, err := loadRunJournal(path)
	if err != nil {
		return RunJournal{}, err
	}
	if journal.State != JournalMutating {
		return RunJournal{}, fmt.Errorf("cannot record mutation intent while journal is %s", journal.State)
	}
	if original.Path == "" {
		return RunJournal{}, fmt.Errorf("captured original path is required")
	}
	if _, _, err := secureTarget(journal.Repository, original.Path); err != nil {
		return RunJournal{}, err
	}
	if err := validateCapturedOriginal(original); err != nil {
		return RunJournal{}, err
	}

	index := -1
	for i := range journal.Originals {
		if journal.Originals[i].Path == original.Path {
			index = i
			break
		}
	}
	if index < 0 {
		// The boundary capture is the authoritative first original. Copy content
		// so callers cannot mutate the durable value after this call returns.
		original.Content = append([]byte(nil), original.Content...)
		journal.Originals = append(journal.Originals, original)
		index = len(journal.Originals) - 1
	}

	sum := sha256.Sum256(intended)
	journal.Originals[index].IntendedSHA256 = hex.EncodeToString(sum[:])
	if err := persistRunJournalAt(path, journal); err != nil {
		return RunJournal{}, err
	}
	return journal, nil
}

func validateCapturedOriginal(original JournalOriginal) error {
	if !original.Existed {
		if original.Mode != 0 || original.SHA256 != "" || len(original.Content) != 0 {
			return fmt.Errorf("captured non-existent original %q contains file metadata", original.Path)
		}
		return nil
	}
	if original.SHA256 == "" {
		return fmt.Errorf("captured original %q is missing sha256", original.Path)
	}
	sum := sha256.Sum256(original.Content)
	actual := hex.EncodeToString(sum[:])
	if original.SHA256 != actual {
		return fmt.Errorf("captured original %q sha256 does not match content", original.Path)
	}
	return nil
}
