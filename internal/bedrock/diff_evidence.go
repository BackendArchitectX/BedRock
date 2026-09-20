package bedrock

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"sort"
)

// DiffHash returns a stable, content-sensitive digest of the mutations BedRock
// actually wrote during this run. It hashes paths plus before/after content
// digests rather than raw content, so persisted evidence does not copy source
// or secrets. An empty change set has no diff hash.
func (c *ChangeSet) DiffHash() string {
	paths := c.ChangedPaths()
	if len(paths) == 0 {
		return ""
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, path := range paths {
		writeHashField(h, []byte(path))
		original := c.originals[path]
		if original.existed {
			h.Write([]byte{1})
			before := sha256.Sum256(original.content)
			writeHashField(h, before[:])
		} else {
			h.Write([]byte{0})
		}
		after := sha256.Sum256(c.written[path])
		writeHashField(h, after[:])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func writeHashField(h hash.Hash, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	h.Write(size[:])
	h.Write(value)
}
