package auditrecord

import "github.com/raystack/frontier/core/auditrecord/models"

// Models moved to a new package to avoid circular dependency with other packages.
// Re-assigned for backward compatibility
type AuditRecord = models.AuditRecord
type Actor = models.Actor
type Resource = models.Resource
type Target = models.Target
type AuditRecordsList = models.AuditRecordsList
type AuditRecordRQLSchema = models.AuditRecordRQLSchema
