package bedrock

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxChangesPerResponse = 32
	maxChangeBytes        = 1024 * 1024
	maxTotalChangeBytes   = 4 * 1024 * 1024
)

type originalFile struct {
	existed bool
	content []byte
	mode    os.FileMode
}

type ChangeSet struct {
	originals map[string]originalFile
	changed   map[string]struct{}
	written   map[string][]byte
}

func NewChangeSet() *ChangeSet {
	return &ChangeSet{
		originals: map[string]originalFile{},
		changed:   map[string]struct{}{},
		written:   map[string][]byte{},
	}
}

func DirtyPaths(root string) (map[string]struct{}, error) {
	dirty := map[string]struct{}{}
	probe := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	probeOut, err := probe.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return nil, fmt.Errorf("inspect git worktree: %w", err)
		}
		return dirty, nil
	}
	if strings.TrimSpace(string(probeOut)) != "true" {
		return dirty, nil
	}

	prefixCmd := exec.Command("git", "-C", root, "rev-parse", "--show-prefix")
	prefixOut, err := prefixCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("inspect git worktree prefix: %w", err)
	}
	prefix := filepath.ToSlash(strings.TrimSuffix(string(prefixOut), "\n"))
	prefix = strings.TrimSuffix(prefix, "\r")

	cmd := exec.Command("git", "-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	parsed, err := parsePorcelainV1Z(out)
	if err != nil {
		return nil, err
	}
	if prefix == "" {
		return parsed, nil
	}
	for path := range parsed {
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		rel := strings.TrimPrefix(path, prefix)
		if rel != "" {
			dirty[rel] = struct{}{}
		}
	}
	return dirty, nil
}

func parsePorcelainV1Z(out []byte) (map[string]struct{}, error) {
	dirty := map[string]struct{}{}
	records := strings.Split(string(out), "\x00")
	for i := 0; i < len(records); i++ {
		record := records[i]
		if record == "" {
			continue
		}
		if len(record) < 4 || record[2] != ' ' {
			return nil, fmt.Errorf("malformed git status record %q", record)
		}
		path := record[3:]
		if path == "" {
			return nil, errors.New("git status returned an empty path")
		}
		dirty[filepath.ToSlash(path)] = struct{}{}

		if record[0] == 'R' || record[0] == 'C' || record[1] == 'R' || record[1] == 'C' {
			i++
			if i >= len(records) || records[i] == "" {
				return nil, fmt.Errorf("git status rename/copy record for %q is missing its source path", path)
			}
			dirty[filepath.ToSlash(records[i])] = struct{}{}
		}
	}
	return dirty, nil
}

func (c *ChangeSet) ExcludeOwnWrites(root string, protected map[string]struct{}) error {
	for rel, expected := range c.written {
		if _, dirty := protected[rel]; !dirty {
			continue
		}
		_, target, err := secureTarget(root, rel)
		if err != nil {
			return err
		}
		current, err := os.ReadFile(target)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("inspect current %s: %w", rel, err)
		}
		if bytes.Equal(current, expected) {
			delete(protected, rel)
		}
	}
	return nil
}

func (c *ChangeSet) Apply(root string, changes []FileChange, protected map[string]struct{}) error {
	if len(changes) > maxChangesPerResponse {
		return fmt.Errorf("provider proposed %d changes; maximum is %d", len(changes), maxChangesPerResponse)
	}
	type prepared struct {
		rel     string
		target  string
		content []byte
		mode    os.FileMode
	}
	seen := map[string]struct{}{}
	total := 0
	preparedChanges := make([]prepared, 0, len(changes))

	for _, change := range changes {
		rel, target, err := secureTarget(root, change.Path)
		if err != nil {
			return err
		}
		if _, duplicate := seen[rel]; duplicate {
			return fmt.Errorf("duplicate change path %q", rel)
		}
		seen[rel] = struct{}{}
		if _, blocked := protected[rel]; blocked {
			return fmt.Errorf("refusing to overwrite pre-existing dirty path %q", rel)
		}
		content := []byte(change.Content)
		if len(content) > maxChangeBytes {
			return fmt.Errorf("change %q exceeds %d bytes", rel, maxChangeBytes)
		}
		total += len(content)
		if total > maxTotalChangeBytes {
			return fmt.Errorf("proposed changes exceed %d total bytes", maxTotalChangeBytes)
		}

		mode := os.FileMode(0o644)
		if info, err := os.Lstat(target); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to write symlink %q", rel)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("refusing to replace non-regular path %q", rel)
			}
			mode = info.Mode().Perm()
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect %q: %w", rel, err)
		}
		preparedChanges = append(preparedChanges, prepared{rel: rel, target: target, content: content, mode: mode})
	}

	for _, p := range preparedChanges {
		if _, captured := c.originals[p.rel]; !captured {
			if data, err := os.ReadFile(p.target); err == nil {
				info, statErr := os.Stat(p.target)
				if statErr != nil {
					_ = c.Rollback(root)
					return statErr
				}
				c.originals[p.rel] = originalFile{existed: true, content: data, mode: info.Mode().Perm()}
			} else if errors.Is(err, os.ErrNotExist) {
				c.originals[p.rel] = originalFile{existed: false}
			} else {
				_ = c.Rollback(root)
				return fmt.Errorf("backup %q: %w", p.rel, err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(p.target), 0o755); err != nil {
			_ = c.Rollback(root)
			return fmt.Errorf("create parent for %q: %w", p.rel, err)
		}
		if err := os.WriteFile(p.target, p.content, p.mode); err != nil {
			_ = c.Rollback(root)
			return fmt.Errorf("write %q: %w", p.rel, err)
		}
		c.changed[p.rel] = struct{}{}
		c.written[p.rel] = append([]byte(nil), p.content...)
	}
	return nil
}

func (c *ChangeSet) ChangedPaths() []string {
	paths := make([]string, 0, len(c.changed))
	for path := range c.changed {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func (c *ChangeSet) Rollback(root string) error {
	var failures []string
	for rel, original := range c.originals {
		_, target, err := secureTarget(root, rel)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		expected, written := c.written[rel]
		if written {
			current, readErr := os.ReadFile(target)
			if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
				failures = append(failures, fmt.Sprintf("inspect current %s: %v", rel, readErr))
				continue
			}
			if readErr == nil && !bytes.Equal(current, expected) {
				failures = append(failures, fmt.Sprintf("refusing to rollback %s because it changed after BedRock wrote it", rel))
				continue
			}
			if errors.Is(readErr, os.ErrNotExist) && original.existed {
				failures = append(failures, fmt.Sprintf("refusing to recreate %s because it was removed after BedRock wrote it", rel))
				continue
			}
		}
		if !original.existed {
			if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
				failures = append(failures, fmt.Sprintf("remove %s: %v", rel, err))
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			failures = append(failures, fmt.Sprintf("mkdir %s: %v", rel, err))
			continue
		}
		if err := os.WriteFile(target, original.content, original.mode); err != nil {
			failures = append(failures, fmt.Sprintf("restore %s: %v", rel, err))
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "; "))
	}
	return nil
}

func secureTarget(root, requested string) (string, string, error) {
	if strings.TrimSpace(requested) == "" {
		return "", "", errors.New("empty change path")
	}
	if filepath.IsAbs(requested) || filepath.VolumeName(requested) != "" {
		return "", "", fmt.Errorf("absolute paths are forbidden: %q", requested)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", "", err
	}
	clean := filepath.Clean(filepath.FromSlash(requested))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes repository: %q", requested)
	}
	relSlash := filepath.ToSlash(clean)
	first := strings.Split(relSlash, "/")[0]
	if first == ".git" || first == ".bedrock" {
		return "", "", fmt.Errorf("protected BedRock path: %q", requested)
	}
	target := filepath.Join(rootAbs, clean)
	checkRel, err := filepath.Rel(rootAbs, target)
	if err != nil || checkRel == ".." || strings.HasPrefix(checkRel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path escapes repository: %q", requested)
	}

	current := rootAbs
	parts := strings.Split(clean, string(filepath.Separator))
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("inspect path %q: %w", requested, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", "", fmt.Errorf("symlink path components are forbidden: %q", requested)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return "", "", fmt.Errorf("non-directory path component in %q", requested)
		}
	}
	return relSlash, target, nil
}
