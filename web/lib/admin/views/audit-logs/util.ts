import { toJsonString } from "@bufbuild/protobuf";
import {
  AuditRecord,
  AuditRecordActor,
  AuditRecordSchema,
} from "@raystack/proton/frontier";

export const getAuditLogActorName = (
  actor?: AuditRecordActor,
  maxLength = 15,
) => {
  if (actor?.type === ACTOR_TYPES.SYSTEM) return "System";

  const name = actor?.title || actor?.name || "-";

  if (actor?.metadata?.["is_super_user"] === true)
    if (name.length > maxLength)
      return name.substring(0, maxLength) + "..." + " (Admin)";
    else return name + " (Admin)";

  return name;
};

const actionBadgeColorPatterns = {
  warning: /invite|unverify|unverified/i,
  success: /success|create|verify|verified/i,
  danger: /error|delete|revoke|remove|disable/i,
};

export const getActionBadgeColor = (action: string) => {
  for (const [color, pattern] of Object.entries(actionBadgeColorPatterns)) {
    if (pattern.test(action)) return color;
  }
  return "accent";
};

export const ACTOR_TYPES = {
  USER: "app/user",
  SERVICE_USER: "app/serviceuser",
  SYSTEM: "system",
} as const;

export const AUDIT_LOG_QUERY_KEY = ["audit-logs", "table-query"];

export const auditLogToJson = (auditLog: AuditRecord) => {
  return toJsonString(AuditRecordSchema, auditLog, { prettySpaces: 2 });
};
