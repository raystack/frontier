package deleter

import (
	"context"
	"maps"
	"slices"

	"github.com/raystack/frontier/core/resource"
)

const metadataDeletable = "deletable"

func (d Service) ProjectResourceBlockers(ctx context.Context, projectID string) ([]Blocker, error) {
	if !d.deletionConfig.Enabled {
		return nil, nil
	}
	resources, err := d.resService.List(ctx, resource.Filter{ProjectID: projectID})
	if err != nil {
		return nil, err
	}
	return d.deletionConfig.blockers(ParentProject, resources), nil
}

// OrganizationResourceBlockers judges the org's own resources and those of
// all its projects by the organization rules.
func (d Service) OrganizationResourceBlockers(ctx context.Context, orgID string) ([]Blocker, error) {
	if !d.deletionConfig.Enabled {
		return nil, nil
	}
	// the org's own resources only; project resources are added below
	resources, err := d.resService.List(ctx, resource.Filter{OrgID: orgID})
	if err != nil {
		return nil, err
	}
	projects, err := d.listOrgProjects(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		projectResources, err := d.resService.List(ctx, resource.Filter{ProjectID: p.ID})
		if err != nil {
			return nil, err
		}
		resources = append(resources, projectResources...)
	}
	return d.deletionConfig.blockers(ParentOrganization, resources), nil
}

func (c Config) blockers(parent Parent, resources []resource.Resource) []Blocker {
	counts := map[string]int{}
	for _, r := range resources {
		if c.blocks(parent, r) {
			counts[r.NamespaceID]++
		}
	}
	var blockers []Blocker
	for _, namespace := range slices.Sorted(maps.Keys(counts)) {
		blockers = append(blockers, Blocker{
			Type:        BlockerResourcesExist,
			Subject:     namespace,
			SubjectType: SubjectResourceNamespace,
			Count:       counts[namespace],
		})
	}
	return blockers
}

func (c Config) blocks(parent Parent, r resource.Resource) bool {
	switch c.Resources[r.NamespaceID][parent] {
	case RuleAllow:
		return false
	case RuleBlockUnlessDeletable:
		deletable, _ := r.Metadata[metadataDeletable].(bool)
		return !deletable
	default:
		// block, or no rule
		return true
	}
}
