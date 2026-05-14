package schema

import (
	"fmt"

	azcore "github.com/authzed/spicedb/pkg/proto/core/v1"
	"github.com/authzed/spicedb/pkg/schemadsl/compiler"
)

// Inheritance is the schema-derived permission lists membership listing needs
// to mirror SpiceDB's project.get / org.project_get chains in Go. Extracted
// from the effective schema at bootstrap so the lists can't drift.
type Inheritance struct {
	// ProjectDirectVisibility: role permissions that make a directly-policied
	// project visible. Matches granted-> arrows of app/project.get.
	ProjectDirectVisibility []string

	// OrganizationToProjectInherit: role permissions on an org policy that
	// expand to all projects in that org. Matches granted-> AND pat_granted->
	// arrows of app/organization.project_get — both walked because PAT
	// all-projects scopes are stored as one pat_granted org policy and need
	// the PAT-only arrows to resolve in the recursive PAT pass.
	OrganizationToProjectInherit []string
}

// ExtractInheritance walks the compiled effective schema. Call after
// ApplyServiceDefinitionOverAZSchema so any custom-resource overlays are in.
func ExtractInheritance(compiled *compiler.CompiledSchema) (Inheritance, error) {
	direct, err := extractInheritanceArrows(compiled, ProjectNamespace, GetPermission)
	if err != nil {
		return Inheritance{}, fmt.Errorf("extract %s.%s arrows: %w", ProjectNamespace, GetPermission, err)
	}
	orgInherit, err := extractInheritanceArrows(compiled, OrganizationNamespace, "project_get")
	if err != nil {
		return Inheritance{}, fmt.Errorf("extract %s.project_get arrows: %w", OrganizationNamespace, err)
	}
	return Inheritance{
		ProjectDirectVisibility:      direct,
		OrganizationToProjectInherit: orgInherit,
	}, nil
}

// extractInheritanceArrows returns the role-relation names on every granted->
// or pat_granted-> arrow under <objectName>.<permissionName>. Errors loudly on
// non-Union rewrites (Intersection/Exclusion would need different handling).
func extractInheritanceArrows(compiled *compiler.CompiledSchema, objectName, permissionName string) ([]string, error) {
	if compiled == nil {
		return nil, fmt.Errorf("compiled schema is nil")
	}

	var def *azcore.NamespaceDefinition
	for _, d := range compiled.ObjectDefinitions {
		if d.GetName() == objectName {
			def = d
			break
		}
	}
	if def == nil {
		return nil, fmt.Errorf("object %q not found in schema", objectName)
	}

	var rel *azcore.Relation
	for _, r := range def.GetRelation() {
		if r.GetName() == permissionName {
			rel = r
			break
		}
	}
	if rel == nil {
		return nil, fmt.Errorf("permission %q not found on %q", permissionName, objectName)
	}

	rewrite := rel.GetUsersetRewrite()
	if rewrite == nil {
		return nil, fmt.Errorf("%s.%s is not a permission (no userset_rewrite)", objectName, permissionName)
	}

	seen := make(map[string]struct{})
	var arrows []string
	if err := collectGrantedArrows(rewrite, objectName, permissionName, &arrows, seen); err != nil {
		return nil, err
	}
	return arrows, nil
}

// collectGrantedArrows walks a UsersetRewrite tree and collects the computed
// userset names from TupleToUserset children whose tupleset relation is
// granted or pat_granted. Recurses into nested rewrites; errors on non-Union.
func collectGrantedArrows(rewrite *azcore.UsersetRewrite, objectName, permissionName string, out *[]string, seen map[string]struct{}) error {
	union := rewrite.GetUnion()
	if union == nil {
		op := "unknown"
		switch {
		case rewrite.GetIntersection() != nil:
			op = "intersection"
		case rewrite.GetExclusion() != nil:
			op = "exclusion"
		}
		return fmt.Errorf("%s.%s uses %s; only union is supported for inheritance extraction", objectName, permissionName, op)
	}

	for _, child := range union.GetChild() {
		switch {
		case child.GetTupleToUserset() != nil:
			ttu := child.GetTupleToUserset()
			tuplesetRel := ttu.GetTupleset().GetRelation()
			if tuplesetRel != RoleGrantRelationName && tuplesetRel != PATGrantRelationName {
				continue
			}
			cu := ttu.GetComputedUserset()
			if cu == nil {
				continue
			}
			permName := cu.GetRelation()
			if permName == "" {
				continue
			}
			if _, ok := seen[permName]; ok {
				continue
			}
			seen[permName] = struct{}{}
			*out = append(*out, permName)
		case child.GetUsersetRewrite() != nil:
			if err := collectGrantedArrows(child.GetUsersetRewrite(), objectName, permissionName, out, seen); err != nil {
				return err
			}
		default:
			// _this, computed_userset, _nil: not granted->/pat_granted-> arrows,
			// nothing to collect.
		}
	}
	return nil
}
