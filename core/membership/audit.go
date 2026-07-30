package membership

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/server/consts"
)

// createAuditRecord writes the audit record and, when the write fails, logs
// every detail of the record so it can be recreated manually.
func (s *Service) createAuditRecord(ctx context.Context, record auditrecord.AuditRecord) {
	if _, err := s.auditRecordRepository.Create(ctx, record); err != nil {
		args := []any{
			"error", err,
			"event", record.Event,
			"org_id", record.OrgID,
			"resource_id", record.Resource.ID,
			"resource_type", record.Resource.Type,
			"resource_name", record.Resource.Name,
			"occurred_at", record.OccurredAt,
		}
		if record.Target != nil {
			args = append(args,
				"target_id", record.Target.ID,
				"target_type", record.Target.Type,
				"target_name", record.Target.Name,
				"target_metadata", record.Target.Metadata,
			)
		}
		// The actor is enriched from the context by the repository, so the
		// failed record carries none; read it from the same place.
		if actorMap, ok := ctx.Value(consts.AuditRecordActorContextKey).(map[string]interface{}); ok {
			if id, ok := actorMap["id"].(string); ok {
				args = append(args, "actor_id", id)
			}
			if actorType, ok := actorMap["type"].(string); ok {
				args = append(args, "actor_type", actorType)
			}
		}
		s.log.WarnContext(ctx, "failed to create audit record", args...)
	}
}

func (s *Service) auditOrgMemberRoleChanged(ctx context.Context, org organization.Organization, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	if err := audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberRoleChangedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id": roleID,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.OrgMemberRoleChangedEvent)
	}
}

func (s *Service) auditOrgMemberAdded(ctx context.Context, org organization.Organization, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})

	if err := audit.GetAuditor(ctx, org.ID).LogWithAttrs(audit.OrgMemberCreatedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id": roleID,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.OrgMemberCreatedEvent)
	}
}

func (s *Service) auditOrgMemberRemoved(ctx context.Context, org organization.Organization, targetID string, targetType pkgAuditRecord.EntityType) {
	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.OrganizationMemberRemovedEvent,
		Resource: auditrecord.Resource{
			ID:   org.ID,
			Type: pkgAuditRecord.OrganizationType,
			Name: org.Title,
		},
		Target: &auditrecord.Target{
			ID:   targetID,
			Type: targetType,
		},
		OrgID:      org.ID,
		OccurredAt: time.Now(),
	})
}

func principalTypeToAuditType(principalType string) (pkgAuditRecord.EntityType, error) {
	switch principalType {
	case schema.ServiceUserPrincipal:
		return pkgAuditRecord.ServiceUserType, nil
	case schema.UserPrincipal:
		return pkgAuditRecord.UserType, nil
	case schema.GroupPrincipal:
		return pkgAuditRecord.GroupType, nil
	case schema.PATPrincipal:
		return pkgAuditRecord.PATType, nil
	default:
		return "", ErrInvalidPrincipalType
	}
}

func (s *Service) auditProjectMember(ctx context.Context, event pkgAuditRecord.Event, prj project.Project, principalID, principalType string, meta map[string]any) {
	targetType, _ := principalTypeToAuditType(principalType)
	if meta == nil {
		meta = map[string]any{}
	}
	meta["principal_type"] = principalType
	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: event,
		Resource: auditrecord.Resource{
			ID:   prj.ID,
			Type: pkgAuditRecord.ProjectType,
			Name: prj.Title,
		},
		Target: &auditrecord.Target{
			ID:       principalID,
			Type:     targetType,
			Metadata: meta,
		},
		OrgID:      prj.Organization.ID,
		OccurredAt: time.Now(),
	})
}

func (s *Service) auditGroupMemberAdded(ctx context.Context, grp group.Group, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberAddedEvent,
		Resource: auditrecord.Resource{
			ID:   grp.ID,
			Type: pkgAuditRecord.GroupType,
			Name: grp.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      grp.OrganizationID,
		OccurredAt: time.Now(),
	})

	if err := audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberCreatedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id":  roleID,
		"group_id": grp.ID,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.GroupMemberCreatedEvent)
	}
}

func (s *Service) auditGroupMemberRoleChanged(ctx context.Context, grp group.Group, p principalInfo, roleID string) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{"role_id": roleID}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberRoleChangedEvent,
		Resource: auditrecord.Resource{
			ID:   grp.ID,
			Type: pkgAuditRecord.GroupType,
			Name: grp.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      grp.OrganizationID,
		OccurredAt: time.Now(),
	})

	if err := audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberRoleChangedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"role_id":  roleID,
		"group_id": grp.ID,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.GroupMemberRoleChangedEvent)
	}
}

func (s *Service) auditGroupMemberRemoved(ctx context.Context, grp group.Group, p principalInfo) {
	targetType, _ := principalTypeToAuditType(p.Type)
	meta := map[string]any{}
	if p.Email != "" {
		meta["email"] = p.Email
	}

	s.createAuditRecord(ctx, auditrecord.AuditRecord{
		Event: pkgAuditRecord.GroupMemberRemovedEvent,
		Resource: auditrecord.Resource{
			ID:   grp.ID,
			Type: pkgAuditRecord.GroupType,
			Name: grp.Title,
		},
		Target: &auditrecord.Target{
			ID:       p.ID,
			Type:     targetType,
			Name:     p.Name,
			Metadata: meta,
		},
		OrgID:      grp.OrganizationID,
		OccurredAt: time.Now(),
	})

	if err := audit.GetAuditor(ctx, grp.OrganizationID).LogWithAttrs(audit.GroupMemberRemovedEvent, audit.Target{
		ID:   p.ID,
		Type: p.Type,
	}, map[string]string{
		"group_id": grp.ID,
	}); err != nil {
		s.log.WarnContext(ctx, "failed to write audit log", "error", err, "event", audit.GroupMemberRemovedEvent)
	}
}
