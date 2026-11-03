import { ConnectError } from '@connectrpc/connect';
import type { DeleteBlocker } from '../../shared/types';

export const ORGANIZATION_DISABLED = 'ORGANIZATION_DISABLED';

export const UNKNOWN_BLOCKER = 'UNKNOWN';

export const GENERIC_DELETE_BLOCKED_MESSAGE =
  'Something is blocking the delete right now. Please try again later or contact support.';

// Words that come from the host app: what it calls an organization
// and what it calls its base plan.
export interface BlockerWording {
  organizationLabel: string;
  basePlanTitle?: string;
}

const DEFAULT_LINES: Record<string, (wording: BlockerWording) => string> = {
  [ORGANIZATION_DISABLED]: ({ organizationLabel }) =>
    `Disabled ${organizationLabel}: Contact your administrator to enable this ${organizationLabel} before deleting it.`,
  ACTIVE_SUBSCRIPTION: ({ organizationLabel, basePlanTitle }) =>
    `Active subscription: Downgrade your subscription to the ${
      basePlanTitle || 'standard plan'
    } before deleting this ${organizationLabel}.`,
  UNPAID_INVOICE: ({ organizationLabel }) =>
    `Outstanding invoice: Pay all outstanding invoices before deleting this ${organizationLabel}.`,
  NEGATIVE_TOKEN_BALANCE: ({ organizationLabel }) =>
    `Negative token balance: Purchase enough tokens to clear your outstanding token balance before deleting this ${organizationLabel}.`
};

export function blockerLines(
  blockers: DeleteBlocker[],
  messages: Record<string, string>,
  wording: BlockerWording
): string[] {
  const lines = new Set<string>();
  let hasUnknown = false;
  for (const { type, subject } of blockers) {
    const line =
      messages[subject] ?? messages[type] ?? DEFAULT_LINES[type]?.(wording);
    if (line) {
      lines.add(line);
    } else {
      hasUnknown = true;
    }
  }
  if (hasUnknown) {
    lines.add(GENERIC_DELETE_BLOCKED_MESSAGE);
  }
  return [...lines];
}

export function toDeleteBlockers(blockers: DeleteBlocker[]): DeleteBlocker[] {
  return blockers.map(({ type, subject, count }) => ({ type, subject, count }));
}

export function blockersFromError(err: ConnectError): DeleteBlocker[] {
  const failures = err.details.filter(
    detail =>
      'type' in detail && detail.type === 'google.rpc.PreconditionFailure'
  );
  if (failures.length === 0) {
    return [{ type: ORGANIZATION_DISABLED, subject: '', count: 0 }];
  }
  const blockers = failures.flatMap(detail => {
    const debug = (detail as { debug?: { violations?: unknown } }).debug;
    const violations = debug?.violations;
    return Array.isArray(violations)
      ? violations.filter(isViolation).map(({ type, subject }) => ({
          type,
          subject: subject ?? '',
          count: 0
        }))
      : [];
  });
  return blockers.length > 0
    ? blockers
    : [{ type: UNKNOWN_BLOCKER, subject: '', count: 0 }];
}

function isViolation(
  violation: unknown
): violation is { type: string; subject?: string } {
  return (
    typeof violation === 'object' &&
    violation !== null &&
    typeof (violation as { type?: unknown }).type === 'string'
  );
}
