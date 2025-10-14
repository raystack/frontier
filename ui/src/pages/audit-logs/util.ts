import { AuditRecordActor } from "@raystack/proton/frontier";

export const isAuditLogActorServiceUser = (actor?: AuditRecordActor) =>
  actor?.type === ACTOR_TYPES.SERVICE_USER;

export const getAuditLogActorName = (actor?: AuditRecordActor) => {
  if (isAuditLogActorServiceUser(actor)) return "System";

  const name = actor?.name || "-";

  if (actor?.metadata?.["is_super_user"] === true) return name + " (Admin)";

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
} as const;
