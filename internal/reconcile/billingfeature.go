package reconcile

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

const KindBillingFeature = "BillingFeature"

type BillingFeatureSpec struct {
	Name     string         `yaml:"name"`
	Title    string         `yaml:"title,omitempty"`
	Metadata map[string]any `yaml:"metadata,omitempty"`
	Delete   bool           `yaml:"delete,omitempty"`
}

type currentBillingFeature struct {
	ID         string
	Name       string
	Title      string
	Metadata   map[string]any
	ProductIDs []string
}

type billingFeatureOp struct {
	action  opAction
	spec    BillingFeatureSpec
	current currentBillingFeature
	detail  string
}

func (o billingFeatureOp) String() string {
	if o.action == opUpdate {
		return fmt.Sprintf("update feature %s (%s)", o.spec.Name, o.detail)
	}
	return fmt.Sprintf("add feature %s", o.spec.Name)
}

func validateBillingFeatureSpec(s BillingFeatureSpec) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("feature name is required")
	}
	if s.Delete {
		return fmt.Errorf("feature %q cannot be deleted: there is no feature delete API; keep the entry in the file", s.Name)
	}
	if _, err := structpb.NewStruct(s.Metadata); err != nil {
		return fmt.Errorf("feature %q metadata cannot be stored: %w", s.Name, err)
	}
	return nil
}

func checkBillingFeatureSpecs(specs []BillingFeatureSpec) error {
	seen := map[string]struct{}{}
	for _, s := range specs {
		if err := validateBillingFeatureSpec(s); err != nil {
			return fmt.Errorf("invalid billing feature spec %q: %w", s.Name, err)
		}
		if _, dup := seen[s.Name]; dup {
			return fmt.Errorf("feature %q is listed more than once", s.Name)
		}
		seen[s.Name] = struct{}{}
	}
	return nil
}

func diffBillingFeatures(desired []BillingFeatureSpec, current []currentBillingFeature) ([]billingFeatureOp, error) {
	if err := checkBillingFeatureSpecs(desired); err != nil {
		return nil, err
	}

	byName := make(map[string]currentBillingFeature, len(current))
	for _, c := range current {
		byName[c.Name] = c
	}

	listed := map[string]struct{}{}
	var adds, updates []billingFeatureOp
	for _, s := range desired {
		cur, exists := byName[s.Name]
		if !exists {
			adds = append(adds, billingFeatureOp{action: opAdd, spec: s})
			continue
		}
		listed[cur.Name] = struct{}{}
		changes, err := billingFeatureChanges(s, cur)
		if err != nil {
			return nil, err
		}
		if len(changes) > 0 {
			updates = append(updates, billingFeatureOp{
				action:  opUpdate,
				spec:    s,
				current: cur,
				detail:  strings.Join(changes, ", "),
			})
		}
	}

	var unlisted []string
	for _, c := range current {
		if _, ok := listed[c.Name]; !ok {
			unlisted = append(unlisted, c.Name)
		}
	}
	if len(unlisted) > 0 {
		sort.Strings(unlisted)
		return nil, fmt.Errorf("features exist on the server but are not in the file: %s; a feature cannot be removed through the API, so add it to the file (frontier export billingfeature lists them all)", strings.Join(unlisted, ", "))
	}

	sortBillingFeatureOpsByName(adds)
	sortBillingFeatureOpsByName(updates)
	return append(adds, updates...), nil
}

func billingFeatureChanges(s BillingFeatureSpec, cur currentBillingFeature) ([]string, error) {
	var changes []string
	if s.Title != cur.Title {
		changes = append(changes, "title")
	}
	desiredMetadata, err := canonicalMetadata(s.Metadata)
	if err != nil {
		return nil, fmt.Errorf("feature %q desired metadata: %w", s.Name, err)
	}
	currentMetadata, err := canonicalMetadata(cur.Metadata)
	if err != nil {
		return nil, fmt.Errorf("feature %q server metadata: %w", s.Name, err)
	}
	if desiredMetadata != currentMetadata {
		changes = append(changes, "metadata")
	}
	return changes, nil
}

func canonicalMetadata(m map[string]any) (string, error) {
	asStored, err := structpb.NewStruct(m)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(asStored.AsMap())
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func sortBillingFeatureOpsByName(ops []billingFeatureOp) {
	sort.Slice(ops, func(i, j int) bool { return ops[i].spec.Name < ops[j].spec.Name })
}
