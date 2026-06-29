// Package reconcile converges Frontier platform resources to a declarative
// desired-state file via the admin API. It is a small framework: each resource
// kind implements Reconciler and registers under its kind, so new kinds
// (roles, plans, products, traits, ...) plug in without changing the command
// or file format. PlatformUser is the first kind.
package reconcile

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Reconciler converges a single resource kind from its desired-state spec.
type Reconciler interface {
	Kind() string
	Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error)
}

// Report summarises what a reconcile did, or would do when dryRun.
type Report struct {
	Kind    string
	DryRun  bool
	Planned []string // human-readable operations (the plan)
	Applied int      // number actually applied (0 when dryRun)
}

// document is one YAML document in a desired-state file: a kind + its spec.
type document struct {
	Kind string    `yaml:"kind"`
	Spec yaml.Node `yaml:"spec"`
}

// Run parses a (possibly multi-document) desired-state file and dispatches each
// document to the registered reconciler for its kind. Documents apply in file
// order; the first error aborts and returns the reports gathered so far.
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
			continue // skip empty documents
		}
		rec, ok := registry[doc.Kind]
		if !ok {
			return reports, fmt.Errorf("no reconciler registered for kind %q", doc.Kind)
		}
		specBytes, err := yaml.Marshal(&doc.Spec)
		if err != nil {
			return reports, fmt.Errorf("marshal spec for kind %q: %w", doc.Kind, err)
		}
		rep, err := rec.Reconcile(ctx, specBytes, dryRun)
		if err != nil {
			return reports, fmt.Errorf("reconcile %s: %w", doc.Kind, err)
		}
		reports = append(reports, rep)
	}
	return reports, nil
}
