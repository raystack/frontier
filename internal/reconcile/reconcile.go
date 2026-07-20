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
// Validate checks a spec without touching the server or applying anything, so
// the whole file can be checked before the first change is made. Export is the
// reverse direction and is part of the contract: it reads the current server
// state and returns it as a spec value, ready to be marshalled into a
// desired-state document. Reconciling an exported document must plan no changes.
type Reconciler interface {
	Kind() string
	Validate(spec []byte) error
	Reconcile(ctx context.Context, spec []byte, dryRun bool) (Report, error)
	Export(ctx context.Context) (spec any, err error)
}

// Report summarises what a reconcile did, or would do when dryRun.
type Report struct {
	Kind    string
	DryRun  bool
	Planned []string // the plan, human-readable
	Applied int      // number actually applied (0 when dryRun)
}

// apiVersion of the document format. A document with no apiVersion is read as
// v1, so files written before the field existed keep working.
const apiVersion = "v1"

type document struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Spec       yaml.Node `yaml:"spec"`
}

type parsedDocument struct {
	kind string
	spec []byte
}

// kindDependencies maps a kind to kinds whose objects it references. In one
// file, a dependency must not come after its dependent: that order would fail
// mid-apply instead of up front.
var kindDependencies = map[string][]string{
	KindRole: {KindPermission},
}

// parseDocuments reads and checks every document in the file before anything
// runs: the version must be known, the kind registered, and the spec present.
// A missing spec marshals to "null"/"" (usually a typo); it is rejected rather
// than treated as an empty list, which for some kinds means "remove everyone".
// Write `spec: []` to mean that on purpose.
func parseDocuments(registry map[string]Reconciler, data []byte) ([]parsedDocument, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var docs []parsedDocument
	for {
		var doc document
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse desired-state: %w", err)
		}
		if doc.Kind == "" {
			// A blank document (a stray "---") is fine to skip, but one that carries
			// content without a kind is a mistake: fail instead of dropping it silently.
			if isBlankDocument(doc) {
				continue
			}
			return nil, fmt.Errorf("a document has content but no kind")
		}
		if doc.APIVersion != "" && doc.APIVersion != apiVersion {
			return nil, fmt.Errorf("document kind %q has unsupported apiVersion %q (want %q)", doc.Kind, doc.APIVersion, apiVersion)
		}
		if _, ok := registry[doc.Kind]; !ok {
			return nil, fmt.Errorf("no reconciler registered for kind %q", doc.Kind)
		}
		specBytes, err := yaml.Marshal(&doc.Spec)
		if err != nil {
			return nil, fmt.Errorf("marshal spec for kind %q: %w", doc.Kind, err)
		}
		if s := strings.TrimSpace(string(specBytes)); s == "" || s == "null" {
			return nil, fmt.Errorf("document kind %q is missing its spec", doc.Kind)
		}
		docs = append(docs, parsedDocument{kind: doc.Kind, spec: specBytes})
	}
	for i, doc := range docs {
		for _, dep := range kindDependencies[doc.kind] {
			for _, later := range docs[i+1:] {
				if later.kind == dep {
					return nil, fmt.Errorf("kind %q must come before kind %q in the file", dep, doc.kind)
				}
			}
		}
	}
	return docs, nil
}

// isBlankDocument reports whether a document carries no content at all (a stray
// "---" separator), as opposed to content that is missing its kind.
func isBlankDocument(doc document) bool {
	if doc.APIVersion != "" {
		return false
	}
	specBytes, err := yaml.Marshal(&doc.Spec)
	if err != nil {
		return false
	}
	s := strings.TrimSpace(string(specBytes))
	return s == "" || s == "null"
}

// decodeSpec unmarshals a kind's spec with unknown fields rejected, so a typo
// in a field name (like `delet: true` for `delete`) fails the plan instead of
// being silently ignored, which would make a run quietly do the wrong thing.
func decodeSpec(spec []byte, out any) error {
	dec := yaml.NewDecoder(bytes.NewReader(spec))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// Run applies a (possibly multi-document) desired-state file. The whole file is
// parsed and checked first, so a malformed later document stops the run before
// anything applies. Documents then dispatch in file order — dependency order is
// the file author's job — and the first error stops the run and returns the
// reports gathered so far.
func Run(ctx context.Context, registry map[string]Reconciler, data []byte, dryRun bool) ([]Report, error) {
	docs, err := parseDocuments(registry, data)
	if err != nil {
		return nil, err
	}
	// Validate every document before applying anything, so a bad entry in a later
	// document stops the run before the first change is made.
	for _, doc := range docs {
		if err := registry[doc.kind].Validate(doc.spec); err != nil {
			return nil, fmt.Errorf("validate %s: %w", doc.kind, err)
		}
	}
	var reports []Report
	for _, doc := range docs {
		rep, err := registry[doc.kind].Reconcile(ctx, doc.spec, dryRun)
		if err != nil {
			// Return rep too: on a partial apply it shows what was applied.
			return append(reports, rep), fmt.Errorf("reconcile %s: %w", doc.kind, err)
		}
		reports = append(reports, rep)
	}
	return reports, nil
}

// Export renders the current server state of one kind as a desired-state YAML
// document that Run accepts as-is.
func Export(ctx context.Context, registry map[string]Reconciler, kind string) ([]byte, error) {
	rec, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("no reconciler registered for kind %q", kind)
	}
	spec, err := rec.Export(ctx)
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", kind, err)
	}
	out, err := yaml.Marshal(struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Spec       any    `yaml:"spec"`
	}{APIVersion: apiVersion, Kind: kind, Spec: spec})
	if err != nil {
		return nil, fmt.Errorf("marshal %s export: %w", kind, err)
	}
	return out, nil
}
