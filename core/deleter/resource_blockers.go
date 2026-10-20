package deleter

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("failed to list resources of project[%s]: %w", projectID, err)
	}
	return d.deletionConfig.blockers(ParentProject, resources), nil
}

func (d Service) OrganizationResourceBlockers(ctx context.Context, orgID string) ([]Blocker, error) {
	if !d.deletionConfig.Enabled {
		return nil, nil
	}
	resources, err := d.resService.List(ctx, resource.Filter{OrgID: orgID})
	if err != nil {
		return nil, fmt.Errorf("failed to list resources of organization[%s]: %w", orgID, err)
	}
	projects, err := d.listOrgProjects(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects of organization[%s]: %w", orgID, err)
	}
	for _, p := range projects {
		projectResources, err := d.resService.List(ctx, resource.Filter{ProjectID: p.ID})
		if err != nil {
			return nil, fmt.Errorf("failed to list resources of project[%s]: %w", p.ID, err)
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
		return true
	}
}
