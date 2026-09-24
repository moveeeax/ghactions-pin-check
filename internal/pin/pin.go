// Package pin classifies GitHub Actions references by how strongly they are
// pinned, and decides whether a reference is acceptable under a given policy.
package pin

import (
	"regexp"
	"strings"
)

// Kind describes what a `uses:` reference is pinned to.
type Kind int

const (
	// Unknown is a reference that could not be classified.
	Unknown Kind = iota
	// SHA is a full 40-hex-character commit SHA — the only immutable pin.
	SHA
	// Tag is a git tag such as "v4" or "v4.1.0" — mutable, can be re-pointed.
	Tag
	// Branch is a branch reference such as "main" — mutable and moves constantly.
	Branch
	// Unpinned is a reference with no version at all (e.g. "actions/checkout").
	Unpinned
	// Local is a local action reference ("./.github/actions/x"); nothing to pin.
	Local
	// Docker is a "docker://" reference; pinning is handled by the image tag/digest.
	Docker
)

func (k Kind) String() string {
	switch k {
	case SHA:
		return "sha"
	case Tag:
		return "tag"
	case Branch:
		return "branch"
	case Unpinned:
		return "unpinned"
	case Local:
		return "local"
	case Docker:
		return "docker"
	default:
		return "unknown"
	}
}

var (
	shaRe      = regexp.MustCompile(`^[0-9a-f]{40}$`)
	tagRe      = regexp.MustCompile(`^v?\d+(\.\d+){0,2}(-[0-9A-Za-z.]+)?$`)
	commonMain = map[string]bool{"main": true, "master": true, "develop": true, "trunk": true}
)

// Ref is the result of classifying a raw `uses:` string.
type Ref struct {
	Raw     string
	Action  string // the "owner/repo" (or path) part, before the '@'
	Version string // the part after '@', empty if none
	Kind    Kind
}

// Classify parses a raw reference such as "actions/checkout@v4" and determines
// how it is pinned.
func Classify(raw string) Ref {
	r := Ref{Raw: raw}

	if strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../") {
		r.Kind = Local
		r.Action = raw
		return r
	}
	if strings.HasPrefix(raw, "docker://") {
		r.Kind = Docker
		r.Action = raw
		return r
	}

	action, version, hasAt := strings.Cut(raw, "@")
	r.Action = action
	r.Version = version

	if !hasAt || version == "" {
		r.Kind = Unpinned
		return r
	}

	switch {
	case shaRe.MatchString(version):
		r.Kind = SHA
	case commonMain[version]:
		r.Kind = Branch
	case tagRe.MatchString(version):
		r.Kind = Tag
	default:
		// Anything else with an '@' that is not a SHA or a version-like tag is
		// treated as a branch/arbitrary ref — still mutable.
		r.Kind = Branch
	}
	return r
}

// Policy decides which reference kinds are acceptable.
type Policy struct {
	// AllowTags permits version tags (e.g. "@v4"). When false, only a full SHA
	// pin is accepted for third-party actions.
	AllowTags bool
}

// OK reports whether the reference satisfies the policy. Local and docker
// references are always acceptable — there is no upstream tag to pin.
func (p Policy) OK(r Ref) bool {
	switch r.Kind {
	case SHA, Local, Docker:
		return true
	case Tag:
		return p.AllowTags
	default:
		return false
	}
}

// Reason returns a short human explanation of why a reference fails the policy.
// It returns "" when the reference is acceptable.
func (p Policy) Reason(r Ref) string {
	if p.OK(r) {
		return ""
	}
	switch r.Kind {
	case Unpinned:
		return "no version specified; pin to a full commit SHA"
	case Branch:
		return "pinned to a mutable branch/ref; pin to a full commit SHA"
	case Tag:
		return "pinned to a mutable tag; pin to a full commit SHA"
	default:
		return "unrecognised reference; pin to a full commit SHA"
	}
}
