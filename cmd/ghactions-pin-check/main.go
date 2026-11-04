// Command ghactions-pin-check lints GitHub Actions workflows for unpinned
// action references. Pinning to a mutable tag or branch is a supply-chain risk:
// the upstream owner (or an attacker who compromises them) can re-point the tag
// to different code that your workflow then runs with your secrets.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/moveeeax/ghactions-pin-check/internal/lint"
	"github.com/moveeeax/ghactions-pin-check/internal/pin"
	"github.com/moveeeax/ghactions-pin-check/internal/report"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "dev"

// resolvedVersion returns the ldflags version when set, otherwise the module
// version recorded by `go install ...@vX.Y.Z`, falling back to "dev".
func resolvedVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("ghactions-pin-check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		path      = fs.String("path", ".", "repository root, workflows directory, or a single workflow file")
		asJSON    = fs.Bool("json", false, "emit machine-readable JSON")
		allowTags = fs.Bool("allow-tags", false, "accept version tags (e.g. @v4) instead of requiring a full commit SHA")
		showVer   = fs.Bool("version", false, "print version and exit")
	)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "ghactions-pin-check — flag unpinned GitHub Actions references\n\n")
		fmt.Fprintf(stderr, "Usage:\n  ghactions-pin-check [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVer {
		fmt.Fprintln(stdout, resolvedVersion())
		return 0
	}

	policy := pin.Policy{AllowTags: *allowTags}
	sum, err := lint.Run(*path, policy)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	if *asJSON {
		if err := report.WriteJSON(stdout, sum); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 2
		}
	} else {
		report.WriteText(stdout, sum)
	}

	if len(sum.Findings) > 0 {
		return 1 // fail CI when anything is unpinned
	}
	return 0
}
