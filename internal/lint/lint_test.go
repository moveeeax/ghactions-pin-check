package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moveeeax/ghactions-pin-check/internal/pin"
)

func writeWorkflow(t *testing.T, root, name, body string) {
	t.Helper()
	dir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunFindsUnpinned(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", `jobs:
  build:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@a81bbbf8298c0fa03ea29cdc473d45769f953675
      - uses: some/action@main
      - uses: bare/action
`)
	sum, err := Run(root, pin.Policy{AllowTags: false})
	if err != nil {
		t.Fatal(err)
	}
	if sum.Refs != 4 {
		t.Errorf("Refs = %d, want 4", sum.Refs)
	}
	// v4 (tag), main (branch), bare (unpinned) fail; the SHA passes.
	if len(sum.Findings) != 3 {
		t.Fatalf("findings = %d, want 3: %+v", len(sum.Findings), sum.Findings)
	}
}

func TestRunAllowTags(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", `jobs:
  build:
    steps:
      - uses: actions/checkout@v4
      - uses: some/action@main
`)
	sum, err := Run(root, pin.Policy{AllowTags: true})
	if err != nil {
		t.Fatal(err)
	}
	// With tags allowed, only the branch pin is a finding.
	if len(sum.Findings) != 1 {
		t.Fatalf("findings = %d, want 1: %+v", len(sum.Findings), sum.Findings)
	}
	if sum.Findings[0].Kind != "branch" {
		t.Errorf("finding kind = %s, want branch", sum.Findings[0].Kind)
	}
}

func TestRunClean(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", `jobs:
  build:
    steps:
      - uses: actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675
      - uses: ./.github/actions/local
`)
	sum, err := Run(root, pin.Policy{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sum.Findings) != 0 {
		t.Errorf("expected no findings, got %+v", sum.Findings)
	}
}
