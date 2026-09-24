// Package lint ties discovery, parsing, classification and reporting together.
package lint

import (
	"github.com/moveeeax/ghactions-pin-check/internal/pin"
	"github.com/moveeeax/ghactions-pin-check/internal/report"
	"github.com/moveeeax/ghactions-pin-check/internal/workflow"
)

// Run scans root, classifies every action reference against policy, and returns
// a Summary of the violations.
func Run(root string, policy pin.Policy) (report.Summary, error) {
	files, err := workflow.Discover(root)
	if err != nil {
		return report.Summary{}, err
	}

	sum := report.Summary{Scanned: len(files)}
	for _, f := range files {
		refs, err := workflow.Refs(f)
		if err != nil {
			return report.Summary{}, err
		}
		for _, r := range refs {
			sum.Refs++
			cl := pin.Classify(r.Raw)
			if reason := policy.Reason(cl); reason != "" {
				sum.Findings = append(sum.Findings, report.New(r, cl, reason))
			}
		}
	}
	return sum, nil
}
