// Package workflow discovers and parses GitHub Actions workflow files, extracting
// every action reference (`uses:`) together with its source location.
package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Ref is a single `uses:` action reference found in a workflow file.
type Ref struct {
	File string // path to the workflow file, relative to the scan root
	Line int    // 1-based line number of the `uses:` entry
	Raw  string // the reference exactly as written, e.g. "actions/checkout@v4"
}

// Discover returns the workflow files under root. When root points at a
// repository, that is the `.github/workflows` directory; when it points
// directly at a directory of YAML files, those are used as-is. A single file
// path is returned verbatim.
func Discover(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{root}, nil
	}

	// Prefer the conventional location when it exists.
	wfDir := filepath.Join(root, ".github", "workflows")
	if fi, err := os.Stat(wfDir); err == nil && fi.IsDir() {
		return yamlFiles(wfDir)
	}
	return yamlFiles(root)
}

func yamlFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			out = append(out, filepath.Join(dir, name))
		}
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil, fmt.Errorf("no workflow files (*.yml/*.yaml) found in %s", dir)
	}
	return out, nil
}

// Refs parses a single workflow file and returns every `uses:` reference.
//
// It is a deliberately line-oriented scanner rather than a full YAML parser:
// `uses:` is always a scalar on its own line, so a targeted scan avoids
// depending on a YAML library and keeps line numbers exact for reporting.
func Refs(path string) ([]Ref, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var refs []Ref
	for i, line := range strings.Split(string(data), "\n") {
		raw, ok := parseUses(line)
		if !ok {
			continue
		}
		refs = append(refs, Ref{File: path, Line: i + 1, Raw: raw})
	}
	return refs, nil
}

// parseUses extracts the action reference from a line containing a `uses:` key.
// It returns ok=false for lines that are comments or that carry no reference.
func parseUses(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	// A leading "- " introduces a list item (e.g. "- uses: ...").
	trimmed = strings.TrimPrefix(trimmed, "- ")
	trimmed = strings.TrimSpace(trimmed)

	if strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	if !strings.HasPrefix(trimmed, "uses:") {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(trimmed, "uses:"))
	value = stripInlineComment(value)
	value = strings.Trim(value, `"'`)
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	return value, true
}

// stripInlineComment removes a trailing "# ..." comment when it is not inside
// quotes. Action references never contain '#', so an unquoted '#' is a comment.
func stripInlineComment(s string) string {
	inSingle, inDouble := false, false
	for i, r := range s {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return s
}
