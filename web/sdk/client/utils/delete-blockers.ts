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

export const ORGANIZATION_DISABLED = 'ORGANIZATION_DISABLED';

const BLOCKERS: Record<
  string,
  {
    label: (wording: BlockerWording) => string;
    instruction: (wording: BlockerWording) => string;
  }
> = {
  ORGANIZATION_DISABLED: {
    label: ({ organizationLabel }) => `Disabled ${organizationLabel}`,
    instruction: ({ organizationLabel }) =>
      `Contact your administrator to enable this ${organizationLabel} before deleting it.`
  },
  ACTIVE_SUBSCRIPTION: {
    label: () => 'Active subscription',
    instruction: ({ organizationLabel, basePlanTitle }) =>
      `Downgrade your subscription to the ${
        basePlanTitle || 'standard plan'
      } before deleting this ${organizationLabel}.`
  },
  UNPAID_INVOICE: {
    label: () => 'Outstanding invoice',
    instruction: ({ organizationLabel }) =>
      `Pay all outstanding invoices before deleting this ${organizationLabel}.`
  },
  NEGATIVE_TOKEN_BALANCE: {
    label: () => 'Negative token balance',
    instruction: ({ organizationLabel }) =>
      `Purchase enough tokens to clear your outstanding token balance before deleting this ${organizationLabel}.`
  }
};

// Used for blocker types this version does not know.
export const GENERIC_DELETE_BLOCKED_MESSAGE =
  'Something is blocking the delete right now. Please try again later or contact support.';

export interface BlockerItem {
  label?: string;
  instruction: string;
}

export function blockerItems(
  blockers: { type: string }[],
  wording: BlockerWording = DEFAULT_WORDING
): BlockerItem[] {
  const items: BlockerItem[] = [];
  let hasUnknown = false;
  for (const type of new Set(blockers.map(blocker => blocker.type))) {
    const blocker = BLOCKERS[type];
    if (blocker) {
      items.push({
        label: blocker.label(wording),
        instruction: blocker.instruction(wording)
      });
    } else {
      hasUnknown = true;
    }
  }
  if (hasUnknown || items.length === 0) {
    items.push({ instruction: GENERIC_DELETE_BLOCKED_MESSAGE });
  }
  return items;
}

export function instructionLines(
  blockers: { type: string }[],
  wording: BlockerWording = DEFAULT_WORDING
): string[] {
  return blockerItems(blockers, wording).map(item => item.instruction);
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
    return instructionLines(blockers, wording).join(' ');
  }
  return GENERIC_DELETE_BLOCKED_MESSAGE;
}
