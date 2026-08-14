package reconcile

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// KindMetaSchema is the desired-state document kind for entity metaschemas.
const KindMetaSchema = "MetaSchema"

// MetaSchemaSpec is one desired metaschema. Name is a built-in metaschema name
// the server knows; Schema is the JSON schema as a string.
type MetaSchemaSpec struct {
	Name   string `yaml:"name"`
	Schema string `yaml:"schema"`
}

// currentMetaSchema is a metaschema as it exists on the server.
type currentMetaSchema struct {
	ID     string
	Name   string
	Schema string
}

// metaSchemaOp is a single planned change. schema is the JSON to write; id is the
// server id for an update, empty when the metaschema must be created. fromDefault
// marks a reset, so the plan can say so.
type metaSchemaOp struct {
	name        string
	id          string
	schema      string
	fromDefault bool
}

func (o metaSchemaOp) String() string {
	switch {
	case o.id == "":
		return fmt.Sprintf("create metaschema %s", o.name)
	case o.fromDefault:
		return fmt.Sprintf("reset metaschema %s to default", o.name)
	default:
		return fmt.Sprintf("set metaschema %s", o.name)
	}
}

// decodeJSON parses a single JSON value, preserving number literals so a large
// or high-precision number is not rounded through float64. It rejects trailing
// data after the value, matching a strict single-document parse.
func decodeJSON(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected trailing data after JSON value")
	}
	return v, nil
}

// canonicalJSON returns a stable form of a JSON document, so two schemas that
// differ only in whitespace or key order compare equal, while a real difference
// in a number is kept (numbers are compared by their literal, not by float64).
// It keeps the export round-trip stable and stops formatting from making a false
// diff.
func canonicalJSON(s string) (string, error) {
	v, err := decodeJSON(s)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// validateMetaSchemaSpecs checks every entry without touching the server: the
// name is a known built-in, no name repeats, and the schema is non-empty valid
// JSON. defaults is the managed set and the source of known names. Names are
// matched case-insensitively, matching the sibling kinds, so `Organization` and
// `organization` both name the same built-in.
//
// It deliberately does NOT require the schema to be a JSON object. The write API
// (CreateMetaSchema/UpdateMetaSchema) accepts any non-empty schema, so a
// non-object schema is a state the server can reach. Rejecting it here would make
// the reconciler stricter than the server, and the export of such a state would
// then fail its own re-reconcile, which breaks the rule 5 round-trip. If a
// non-object schema should be rejected, that check belongs on the write API,
// where it also stops the state from being reached in the first place.
func validateMetaSchemaSpecs(specs []MetaSchemaSpec, defaults map[string]string) error {
	seen := map[string]struct{}{}
	for _, s := range specs {
		name := strings.ToLower(strings.TrimSpace(s.Name))
		if name == "" {
			return fmt.Errorf("metaschema name is required")
		}
		if _, ok := defaults[name]; !ok {
			return fmt.Errorf("unknown metaschema %q", s.Name)
		}
		if _, dup := seen[name]; dup {
			return fmt.Errorf("metaschema %q is listed more than once", name)
		}
		seen[name] = struct{}{}
		if strings.TrimSpace(s.Schema) == "" {
			return fmt.Errorf("metaschema %q: schema is required", name)
		}
		if _, err := canonicalJSON(s.Schema); err != nil {
			return fmt.Errorf("metaschema %q: schema is not valid JSON: %w", name, err)
		}
	}
	return nil
}

// diffMetaSchemas returns the ops that make the server's built-in metaschemas
// match the desired spec. The file is the full desired state: a built-in the file
// lists is set to its schema, and a built-in the file leaves out is reset to its
// shipped default. defaults is the managed set: its keys are the built-ins, its
// values the reset targets.
func diffMetaSchemas(desired []MetaSchemaSpec, current []currentMetaSchema, defaults map[string]string) ([]metaSchemaOp, error) {
	desiredByName := make(map[string]string, len(desired))
	for _, s := range desired {
		desiredByName[strings.ToLower(strings.TrimSpace(s.Name))] = s.Schema
	}
	currentByName := make(map[string]currentMetaSchema, len(current))
	for _, c := range current {
		currentByName[c.Name] = c
	}

	names := make([]string, 0, len(defaults))
	for name := range defaults {
		names = append(names, name)
	}
	sort.Strings(names)

	var ops []metaSchemaOp
	for _, name := range names {
		want, inFile := desiredByName[name]
		if !inFile {
			want = defaults[name]
		}
		wantCanon, err := canonicalJSON(want)
		if err != nil {
			return nil, fmt.Errorf("metaschema %q: schema is not valid JSON: %w", name, err)
		}

		cur, exists := currentByName[name]
		if exists {
			if curCanon, err := canonicalJSON(cur.Schema); err == nil && curCanon == wantCanon {
				continue
			}
		}
		ops = append(ops, metaSchemaOp{
			name:        name,
			id:          cur.ID, // empty when the metaschema is not on the server
			schema:      want,
			fromDefault: !inFile,
		})
	}
	return ops, nil
}

// exportMetaSchemas returns the built-ins whose current schema differs from the
// default, sorted by name, as a desired-state spec. A built-in at its default is
// omitted, so reconciling an export plans no changes.
func exportMetaSchemas(current []currentMetaSchema, defaults map[string]string) ([]MetaSchemaSpec, error) {
	byName := make(map[string]currentMetaSchema, len(current))
	for _, c := range current {
		byName[c.Name] = c
	}
	names := make([]string, 0, len(defaults))
	for name := range defaults {
		names = append(names, name)
	}
	sort.Strings(names)

	var specs []MetaSchemaSpec
	for _, name := range names {
		cur, exists := byName[name]
		if !exists {
			continue
		}
		curCanon, err := canonicalJSON(cur.Schema)
		if err != nil {
			// A stored schema that is not valid JSON still round-trips as its own
			// literal, so emit it rather than dropping it.
			specs = append(specs, MetaSchemaSpec{Name: name, Schema: cur.Schema})
			continue
		}
		if defCanon, err := canonicalJSON(defaults[name]); err == nil && defCanon == curCanon {
			continue
		}
		specs = append(specs, MetaSchemaSpec{Name: name, Schema: cur.Schema})
	}
	return specs, nil
}
