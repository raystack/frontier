import { ConnectError } from '@connectrpc/connect';

// The server reports why an organization cannot be deleted as a list of
// blockers, each with a machine-readable type. These are the types the
// server knows today, mapped to a short instruction the user can act on.
// The wording must stay in line with the server behavior: a paid
// subscription must be downgraded, unpaid invoices must be paid, and a
// token debt is settled through support.
const BLOCKER_INSTRUCTIONS: Record<string, (count: number) => string> = {
  ACTIVE_SUBSCRIPTION: () =>
    'Please downgrade the subscription to the standard plan',
  UNPAID_INVOICE: count =>
    count > 1
      ? `Please pay the ${count} open invoices from the billing page`
      : 'Please pay the open invoice from the billing page',
  NEGATIVE_TOKEN_BALANCE: () =>
    'Your token balance is negative. Please add tokens from the Tokens page to bring it back up, or pay the invoice when it arrives, and then you can delete'
};

// Shown when the server reports a blocker kind this version does not know,
// or when the error could not be read at all.
export const GENERIC_DELETE_BLOCKED_MESSAGE =
  'Something is blocking the delete right now. Please try again later or contact support.';

// instructionLines turns a list of blockers into one instruction per kind
// of blocker. Blockers of the same kind are counted so the instruction can
// say "the 2 open invoices". A kind without a known instruction becomes the
// generic message, once.
export function instructionLines(blockers: { type: string }[]): string[] {
  const counts = new Map<string, number>();
  for (const blocker of blockers) {
    counts.set(blocker.type, (counts.get(blocker.type) ?? 0) + 1);
  }
  const lines: string[] = [];
  let hasUnknown = false;
  for (const [type, count] of counts) {
    const instruction = BLOCKER_INSTRUCTIONS[type];
    if (instruction) {
      lines.push(instruction(count));
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
export function deleteBlockedDescription(err: ConnectError): string {
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
    const lines = instructionLines(blockers);
    if (lines.length > 0) {
      return lines.join('. ');
    }
  }
  return GENERIC_DELETE_BLOCKED_MESSAGE;
}
