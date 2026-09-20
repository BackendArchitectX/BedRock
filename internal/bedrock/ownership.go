package bedrock

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

// ValidateOwnWrites ensures files written by an earlier attempt still contain
// exactly the bytes BedRock wrote. A missing or changed path means another
// actor has taken ownership, even when Git considers the resulting state clean.
func (c *ChangeSet) ValidateOwnWrites(root string) error {
	for rel, expected := range c.written {
		_, target, err := secureTarget(root, rel)
		if err != nil {
			return err
		}
		current, err := os.ReadFile(target)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("refusing to continue because %s was removed after BedRock wrote it", rel)
			}
			return fmt.Errorf("inspect current %s: %w", rel, err)
		}
		if !bytes.Equal(current, expected) {
			return fmt.Errorf("refusing to continue because %s changed after BedRock wrote it", rel)
		}
	}
	return nil
}
