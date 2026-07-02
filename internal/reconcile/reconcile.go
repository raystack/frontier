// Package reconcile makes Frontier platform resources match a desired-state file
// through the admin API. Each resource kind implements Reconciler and registers
// under its kind, so new kinds plug in without changing the command or file
// format. PlatformUser is the first kind.
package reconcile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Reconciler makes a single resource kind match its desired-state spec.
type Reconciler interface {
	Kind() string
	Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error)
}

// Report summarises what a reconcile did, or would do when dryRun.
type Report struct {
	Kind    string
	DryRun  bool
	Planned []string // the plan, human-readable
	Applied int      // number actually applied (0 when dryRun)
}

type document struct {
	Kind string    `yaml:"kind"`
	Spec yaml.Node `yaml:"spec"`
}

// Run dispatches each document in a (possibly multi-document) desired-state file
// to the reconciler for its kind, in file order. The first error stops the run
// and returns the reports gathered so far.
func Run(ctx context.Context, registry map[string]Reconciler, data []byte, dryRun bool) ([]Report, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var reports []Report
	for {
		var doc document
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return reports, fmt.Errorf("parse desired-state: %w", err)
		}
		if doc.Kind == "" {
			continue
		}
		rec, ok := registry[doc.Kind]
		if !ok {
			return reports, fmt.Errorf("no reconciler registered for kind %q", doc.Kind)
		}
		specBytes, err := yaml.Marshal(&doc.Spec)
		if err != nil {
			return reports, fmt.Errorf("marshal spec for kind %q: %w", doc.Kind, err)
		}
		// A missing spec marshals to "null"/"" (usually a typo). Reject it rather
		// than treat it as an empty list that removes everyone; write `spec: []` to
		// mean that on purpose.
		if s := strings.TrimSpace(string(specBytes)); s == "" || s == "null" {
			return reports, fmt.Errorf("document kind %q is missing its spec", doc.Kind)
		}
		rep, err := rec.Reconcile(ctx, specBytes, dryRun)
		if err != nil {
			// Return rep too: on a partial apply it shows what was applied.
			return append(reports, rep), fmt.Errorf("reconcile %s: %w", doc.Kind, err)
		}
		reports = append(reports, rep)
	}
	return reports, nil
}
