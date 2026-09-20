package bedrock

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type candidateFile struct {
	path    string
	content string
	score   int
}

var ignoredDirs = map[string]struct{}{
	".git": {}, ".bedrock": {}, ".ssh": {}, ".aws": {}, ".azure": {}, ".docker": {}, ".gnupg": {}, ".kube": {},
	"node_modules": {}, "vendor": {}, "dist": {}, "build": {}, "target": {}, ".idea": {}, ".vscode": {},
}

func Snapshot(root, task string, maxFiles, maxBytes int) ([]FileContext, error) {
	if maxFiles <= 0 {
		maxFiles = 80
	}
	if maxBytes <= 0 {
		maxBytes = 200 * 1024
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	eligible, gitRepo, err := gitContextPaths(root)
	if err != nil {
		return nil, err
	}
	terms := taskTerms(task)
	const maxScannedBytes = 4 * 1024 * 1024
	const maxPerFile = 64 * 1024
	const maxCandidates = 500

	var candidates []candidateFile
	scanned := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			if _, skip := ignoredDirs[d.Name()]; skip {
				return filepath.SkipDir
			}
			return nil
		}
		if len(candidates) >= maxCandidates || scanned >= maxScannedBytes {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if gitRepo {
			if _, ok := eligible[rel]; !ok {
				return nil
			}
		}
		if sensitiveContextPath(d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil || !info.Mode().IsRegular() || info.Size() > maxPerFile {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if bytes.IndexByte(data, 0) >= 0 {
			return nil
		}
		if scanned+len(data) > maxScannedBytes {
			return nil
		}
		scanned += len(data)
		text := string(data)
		candidates = append(candidates, candidateFile{
			path:    rel,
			content: text,
			score:   contextScore(rel, text, terms),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan repository: %w", err)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].path < candidates[j].path
	})

	result := make([]FileContext, 0, min(maxFiles, len(candidates)))
	used := 0
	for _, c := range candidates {
		if len(result) >= maxFiles {
			break
		}
		size := len(c.path) + len(c.content)
		if used+size > maxBytes {
			continue
		}
		result = append(result, FileContext{Path: c.path, Content: c.content})
		used += size
	}
	return result, nil
}

func gitContextPaths(root string) (map[string]struct{}, bool, error) {
	probe := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	probeOut, err := probe.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return nil, false, fmt.Errorf("inspect git worktree: %w", err)
		}
		return nil, false, nil
	}
	if strings.TrimSpace(string(probeOut)) != "true" {
		return nil, false, nil
	}

	cmd := exec.Command("git", "-C", root, "ls-files", "-co", "--exclude-standard", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, true, fmt.Errorf("list git context files: %w", err)
	}
	paths := make(map[string]struct{})
	for _, raw := range bytes.Split(out, []byte{0}) {
		if len(raw) == 0 {
			continue
		}
		paths[filepath.ToSlash(string(raw))] = struct{}{}
	}
	return paths, true, nil
}

func sensitiveContextPath(name string) bool {
	name = strings.ToLower(name)
	if name == ".env" || name == ".envrc" || strings.HasPrefix(name, ".env.") {
		return true
	}
	switch name {
	case ".netrc", ".npmrc", ".pypirc", ".git-credentials", "credentials", "credentials.json",
		"application_default_credentials.json", "service-account.json", "secrets.json",
		"secrets.yaml", "secrets.yml", "id_rsa", "id_dsa", "id_ecdsa", "id_ed25519":
		return true
	}
	for _, suffix := range []string{".pem", ".key", ".p12", ".pfx", ".jks", ".keystore", ".tfstate", ".tfstate.backup"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func taskTerms(task string) []string {
	seen := map[string]struct{}{}
	var terms []string
	for _, part := range strings.FieldsFunc(strings.ToLower(task), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-'
	}) {
		if len(part) < 3 {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		terms = append(terms, part)
	}
	return terms
}

func contextScore(path, content string, terms []string) int {
	p := strings.ToLower(path)
	c := strings.ToLower(content)
	score := 0
	base := strings.ToLower(filepath.Base(path))
	switch base {
	case "readme.md", "go.mod", "package.json", "pom.xml", "pyproject.toml", "cargo.toml":
		score += 5
	}
	for _, term := range terms {
		if strings.Contains(p, term) {
			score += 8
		}
		if strings.Contains(c, term) {
			score += 2
		}
	}
	return score
}
