import { ConnectError } from '@connectrpc/connect';

// Words that come from the host app: what it calls an organization
// and what it calls its base plan.
export interface BlockerWording {
  organizationLabel: string;
  basePlanTitle?: string;
}

const DEFAULT_WORDING: BlockerWording = {
  organizationLabel: 'organization'
};

// One sentence per blocker type the server can return.
const BLOCKER_INSTRUCTIONS: Record<
  string,
  (count: number, wording: BlockerWording) => string
> = {
  ACTIVE_SUBSCRIPTION: (_, { organizationLabel, basePlanTitle }) =>
    `Downgrade your current subscription and switch to the ${
      basePlanTitle || 'standard plan'
    } before deleting this ${organizationLabel}.`,
  UNPAID_INVOICE: (count, { organizationLabel }) =>
    count > 1
      ? `Please pay your outstanding invoices before deleting this ${organizationLabel}.`
      : `Please pay your outstanding invoice before deleting this ${organizationLabel}.`,
  NEGATIVE_TOKEN_BALANCE: (_, { organizationLabel }) =>
    `Please purchase enough tokens to clear your outstanding token balance before deleting this ${organizationLabel}.`
};

// Used for blocker types this version does not know.
export const GENERIC_DELETE_BLOCKED_MESSAGE =
  'Something is blocking the delete right now. Please try again later or contact support.';

// Returns one sentence per blocker type.
export function instructionLines(
  blockers: { type: string }[],
  wording: BlockerWording = DEFAULT_WORDING
): string[] {
  const counts = new Map<string, number>();
  for (const blocker of blockers) {
    counts.set(blocker.type, (counts.get(blocker.type) ?? 0) + 1);
  }
  const lines: string[] = [];
  let hasUnknown = false;
  for (const [type, count] of counts) {
    const instruction = BLOCKER_INSTRUCTIONS[type];
    if (instruction) {
      lines.push(instruction(count, wording));
    } else {
      hasUnknown = true;
    }
  }
  if (hasUnknown) {
    lines.push(GENERIC_DELETE_BLOCKED_MESSAGE);
  }
  return lines;
}

// Reads the blockers from a failed_precondition error and returns the
// sentences as one string. The server sends them as a
// google.rpc.PreconditionFailure detail, which Connect exposes in the
// detail's debug field.
export function deleteBlockedDescription(
  err: ConnectError,
  wording: BlockerWording = DEFAULT_WORDING
): string {
  for (const detail of err.details) {
    if (!('type' in detail) || detail.type !== 'google.rpc.PreconditionFailure') {
      continue;
    }
    const debug = (detail as { debug?: unknown }).debug;
    if (typeof debug !== 'object' || debug === null) {
      continue;
    }
    const violations = (debug as { violations?: unknown }).violations;
    if (!Array.isArray(violations)) {
      continue;
    }
    const blockers = violations
      .filter(
        (violation): violation is { type: string } =>
          typeof violation === 'object' &&
          violation !== null &&
          typeof (violation as { type?: unknown }).type === 'string'
      );
    const lines = instructionLines(blockers, wording);
    if (lines.length > 0) {
      return lines.join(' ');
    }
  }
  return GENERIC_DELETE_BLOCKED_MESSAGE;
}
