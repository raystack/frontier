import { ConnectError } from '@connectrpc/connect';

// Words the instructions are built from. The host app decides what an
// organization is called (for example "workspace") and what its base plan
// is called (for example "Standard Plan"), so both come from the caller.
export interface BlockerWording {
  // The organization noun in lower case, as used mid-sentence.
  organizationLabel: string;
  // The display title of the plan a paid subscription must be moved to.
  // Falls back to a plain "standard plan" when the host app set none.
  basePlanTitle?: string;
}

const DEFAULT_WORDING: BlockerWording = {
  organizationLabel: 'organization'
};

// The server reports why an organization cannot be deleted as a list of
// blockers, each with a machine-readable type. These are the types the
// server knows today, mapped to one full sentence the user can act on.
// The wording must stay in line with the server behavior: a paid
// subscription must be downgraded, unpaid invoices must be paid, and a
// token debt is cleared by buying tokens.
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
      ? `Please pay your ${count} outstanding invoices before deleting this ${organizationLabel}.`
      : `Please pay your outstanding invoice before deleting this ${organizationLabel}.`,
  NEGATIVE_TOKEN_BALANCE: (_, { organizationLabel }) =>
    `Please purchase enough tokens to clear your outstanding token balance before deleting this ${organizationLabel}.`
};

// Shown when the server reports a blocker kind this version does not know,
// or when the error could not be read at all.
export const GENERIC_DELETE_BLOCKED_MESSAGE =
  'Something is blocking the delete right now. Please try again later or contact support.';

// instructionLines turns a list of blockers into one sentence per kind of
// blocker. Blockers of the same kind are counted so the sentence can say
// "your 2 outstanding invoices". A kind without a known instruction becomes
// the generic message, once.
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

// deleteBlockedDescription reads the blockers out of a failed_precondition
// error from DeleteOrganization and returns the instructions as one string.
// The server attaches them as a google.rpc.PreconditionFailure detail; over
// the Connect JSON protocol that detail arrives with a ready-made JSON copy
// in its debug field. When the error carries nothing readable, the generic
// message is returned, so the raw server text is never shown.
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
