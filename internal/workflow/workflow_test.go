package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseUses(t *testing.T) {
	cases := []struct {
		line string
		want string
		ok   bool
	}{
		{"      - uses: actions/checkout@v4", "actions/checkout@v4", true},
		{"        uses: actions/setup-go@v5 # comment", "actions/setup-go@v5", true},
		{`    uses: "owner/action@main"`, "owner/action@main", true},
		{"    uses: 'owner/action@v1'", "owner/action@v1", true},
		{"      # uses: actions/checkout@v4", "", false},
		{"      run: echo uses: not-an-action", "", false},
		{"      uses:", "", false},
		{"      name: build", "", false},
	}
	for _, c := range cases {
		got, ok := parseUses(c.line)
		if ok != c.ok || got != c.want {
			t.Errorf("parseUses(%q) = (%q,%v), want (%q,%v)", c.line, got, ok, c.want, c.ok)
		}
	}
}

func TestRefs(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "ci.yml")
	content := `name: ci
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@a81bbbf8298c0fa03ea29cdc473d45769f953675
      - name: test
        run: go test ./...
`
	if err := os.WriteFile(wf, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	refs, err := Refs(wf)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}
	if refs[0].Raw != "actions/checkout@v4" || refs[0].Line != 7 {
		t.Errorf("ref[0] = %+v, want checkout@v4 at line 7", refs[0])
	}
	if refs[1].Line != 8 {
		t.Errorf("ref[1] line = %d, want 8", refs[1].Line)
	}
}

func TestDiscoverPrefersGithubWorkflows(t *testing.T) {
	root := t.TempDir()
	wfDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"a.yml", "b.yaml"} {
		if err := os.WriteFile(filepath.Join(wfDir, n), []byte("on: push\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A stray YAML at the repo root must NOT be picked up.
	if err := os.WriteFile(filepath.Join(root, "config.yml"), []byte("x: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2 (%v)", len(files), files)
	}
}

func TestDiscoverSingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "one.yml")
	if err := os.WriteFile(f, []byte("on: push\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := Discover(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != f {
		t.Errorf("Discover(file) = %v, want [%s]", files, f)
	}
}

func TestDiscoverEmpty(t *testing.T) {
	dir := t.TempDir()
	if _, err := Discover(dir); err == nil {
		t.Error("expected an error scanning a directory with no workflow files")
	}
}
