import { Code, ConnectError } from '@connectrpc/connect';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  ORGANIZATION_DISABLED,
  blockerItems,
  deleteBlockedDescription,
  instructionLines
} from './delete-blockers';

const wording = { organizationLabel: 'workspace', basePlanTitle: 'Standard Plan' };

describe('blockerItems', () => {
  it('gives a label and one full sentence per kind of blocker', () => {
    const items = blockerItems(
      [
        { type: 'ACTIVE_SUBSCRIPTION' },
        { type: 'UNPAID_INVOICE' },
        { type: 'NEGATIVE_TOKEN_BALANCE' }
      ],
      wording
    );
    expect(items).toEqual([
      {
        label: 'Active subscription',
        instruction:
          'Downgrade your subscription to the Standard Plan before deleting this workspace.'
      },
      {
        label: 'Outstanding invoice',
        instruction:
          'Pay all outstanding invoices before deleting this workspace.'
      },
      {
        label: 'Negative token balance',
        instruction:
          'Purchase enough tokens to clear your outstanding token balance before deleting this workspace.'
      }
    ]);
  });

  it('names the host app terminology for a disabled organization', () => {
    expect(blockerItems([{ type: ORGANIZATION_DISABLED }], wording)).toEqual([
      {
        label: 'Disabled workspace',
        instruction:
          'Contact your administrator to enable this workspace before deleting it.'
      }
    ]);
  });

  it('shows blockers of the same kind once', () => {
    const items = blockerItems(
      [{ type: 'UNPAID_INVOICE' }, { type: 'UNPAID_INVOICE' }],
      wording
    );
    expect(items).toHaveLength(1);
  });

  it('adds the generic message once, with no label, for unknown kinds', () => {
    const items = blockerItems(
      [
        { type: 'UNPAID_INVOICE' },
        { type: 'SOMETHING_NEW' },
        { type: 'SOMETHING_ELSE' }
      ],
      wording
    );
    expect(items).toEqual([
      {
        label: 'Outstanding invoice',
        instruction:
          'Pay all outstanding invoices before deleting this workspace.'
      },
      { instruction: GENERIC_DELETE_BLOCKED_MESSAGE }
    ]);
  });

  it('gives the generic message when there are no blockers', () => {
    expect(blockerItems([], wording)).toEqual([
      { instruction: GENERIC_DELETE_BLOCKED_MESSAGE }
    ]);
  });
});

describe('instructionLines', () => {
  it('falls back to plain words when the host app set no terminology', () => {
    const lines = instructionLines([{ type: 'ACTIVE_SUBSCRIPTION' }]);
    expect(lines).toEqual([
      'Downgrade your subscription to the standard plan before deleting this organization.'
    ]);
  });

  it('shows the generic message once for unknown kinds', () => {
    const lines = instructionLines(
      [{ type: 'SOMETHING_NEW' }, { type: 'SOMETHING_ELSE' }],
      wording
    );
    expect(lines).toEqual([GENERIC_DELETE_BLOCKED_MESSAGE]);
  });
});

describe('deleteBlockedDescription', () => {
  it('reads the blockers out of the precondition failure detail', () => {
    const err = new ConnectError('blocked', Code.FailedPrecondition);
    err.details.push({
      type: 'google.rpc.PreconditionFailure',
      value: new Uint8Array(),
      debug: {
        violations: [
          { type: 'ACTIVE_SUBSCRIPTION', subject: 'sub-1' },
          { type: 'NEGATIVE_TOKEN_BALANCE', subject: 'acct-1' }
        ]
      }
    });
    expect(deleteBlockedDescription(err, wording)).toBe(
      'Downgrade your subscription to the Standard Plan before deleting this workspace. ' +
        'Purchase enough tokens to clear your outstanding token balance before deleting this workspace.'
    );
  });

  it('returns the generic message when the error carries no blockers', () => {
    const err = new ConnectError('blocked', Code.FailedPrecondition);
    expect(deleteBlockedDescription(err, wording)).toBe(
      GENERIC_DELETE_BLOCKED_MESSAGE
    );
  });
});
