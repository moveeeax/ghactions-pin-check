// Package report renders lint findings in human and machine-readable form.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/moveeeax/ghactions-pin-check/internal/pin"
	"github.com/moveeeax/ghactions-pin-check/internal/workflow"
)

// Finding is one unpinned (policy-violating) action reference.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Action  string `json:"action"`
	Version string `json:"version,omitempty"`
	Kind    string `json:"kind"`
	Raw     string `json:"raw"`
	Reason  string `json:"reason"`
}

// New builds a Finding from a workflow reference and its classification.
func New(ref workflow.Ref, cl pin.Ref, reason string) Finding {
	return Finding{
		File:    ref.File,
		Line:    ref.Line,
		Action:  cl.Action,
		Version: cl.Version,
		Kind:    cl.Kind.String(),
		Raw:     cl.Raw,
		Reason:  reason,
	}
}

// Summary is the full lint result.
type Summary struct {
	Scanned  int       `json:"files_scanned"`
	Refs     int       `json:"refs_checked"`
	Findings []Finding `json:"findings"`
}

// WriteText renders a Summary as grouped, human-readable text.
func WriteText(w io.Writer, s Summary) {
	if len(s.Findings) == 0 {
		fmt.Fprintf(w, "OK — %d reference(s) across %d file(s) are all pinned.\n", s.Refs, s.Scanned)
		return
	}

	byFile := map[string][]Finding{}
	var order []string
	for _, f := range s.Findings {
		if _, seen := byFile[f.File]; !seen {
			order = append(order, f.File)
		}
		byFile[f.File] = append(byFile[f.File], f)
	}

	for _, file := range order {
		fmt.Fprintf(w, "%s\n", filepath.ToSlash(file))
		for _, f := range byFile[file] {
			fmt.Fprintf(w, "  line %d: %s\n", f.Line, f.Raw)
			fmt.Fprintf(w, "          %s (%s)\n", f.Reason, f.Kind)
		}
	}
	fmt.Fprintf(w, "\n%d unpinned reference(s) in %d file(s) (%d checked).\n",
		len(s.Findings), len(order), s.Refs)
}

// WriteJSON renders a Summary as indented JSON.
func WriteJSON(w io.Writer, s Summary) error {
	if s.Findings == nil {
		s.Findings = []Finding{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
