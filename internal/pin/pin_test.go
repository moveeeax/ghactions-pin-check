package pin

import "testing"

func TestClassify(t *testing.T) {
	sha := "a81bbbf8298c0fa03ea29cdc473d45769f953675"
	cases := []struct {
		raw     string
		wantK   Kind
		action  string
		version string
	}{
		{"actions/checkout@" + sha, SHA, "actions/checkout", sha},
		{"actions/checkout@v4", Tag, "actions/checkout", "v4"},
		{"actions/checkout@v4.1.0", Tag, "actions/checkout", "v4.1.0"},
		{"actions/checkout@1.2.3", Tag, "actions/checkout", "1.2.3"},
		{"actions/checkout@main", Branch, "actions/checkout", "main"},
		{"owner/action@some-feature-branch", Branch, "owner/action", "some-feature-branch"},
		{"actions/checkout", Unpinned, "actions/checkout", ""},
		{"actions/checkout@", Unpinned, "actions/checkout", ""},
		{"./.github/actions/local", Local, "./.github/actions/local", ""},
		{"docker://alpine:3.20", Docker, "docker://alpine:3.20", ""},
	}
	for _, c := range cases {
		got := Classify(c.raw)
		if got.Kind != c.wantK {
			t.Errorf("Classify(%q).Kind = %v, want %v", c.raw, got.Kind, c.wantK)
		}
		if got.Action != c.action {
			t.Errorf("Classify(%q).Action = %q, want %q", c.raw, got.Action, c.action)
		}
		if got.Version != c.version {
			t.Errorf("Classify(%q).Version = %q, want %q", c.raw, got.Version, c.version)
		}
	}
}

func TestPolicyOK(t *testing.T) {
	sha := "a81bbbf8298c0fa03ea29cdc473d45769f953675"
	strict := Policy{AllowTags: false}
	lax := Policy{AllowTags: true}

	if !strict.OK(Classify("actions/checkout@" + sha)) {
		t.Error("strict policy should accept a SHA pin")
	}
	if strict.OK(Classify("actions/checkout@v4")) {
		t.Error("strict policy must reject a tag")
	}
	if !lax.OK(Classify("actions/checkout@v4")) {
		t.Error("lax policy should accept a tag")
	}
	if lax.OK(Classify("actions/checkout@main")) {
		t.Error("even lax policy must reject a branch")
	}
	if lax.OK(Classify("actions/checkout")) {
		t.Error("even lax policy must reject an unpinned ref")
	}
	if !strict.OK(Classify("./.github/actions/x")) {
		t.Error("local action should always be accepted")
	}
	if !strict.OK(Classify("docker://alpine:3.20")) {
		t.Error("docker ref should always be accepted")
	}
}

func TestReason(t *testing.T) {
	strict := Policy{}
	if r := strict.Reason(Classify("actions/checkout@v4")); r == "" {
		t.Error("expected a non-empty reason for a rejected tag")
	}
	sha := "a81bbbf8298c0fa03ea29cdc473d45769f953675"
	if r := strict.Reason(Classify("actions/checkout@" + sha)); r != "" {
		t.Errorf("expected empty reason for an accepted SHA, got %q", r)
	}
}
