import { Code, ConnectError } from '@connectrpc/connect';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  deleteBlockedDescription,
  instructionLines
} from './delete-blockers';

const wording = { organizationLabel: 'workspace', basePlanTitle: 'Standard Plan' };

describe('instructionLines', () => {
  it('writes one full sentence per kind of blocker', () => {
    const lines = instructionLines(
      [
        { type: 'ACTIVE_SUBSCRIPTION' },
        { type: 'UNPAID_INVOICE' },
        { type: 'NEGATIVE_TOKEN_BALANCE' }
      ],
      wording
    );
    expect(lines).toEqual([
      'Downgrade your current subscription and switch to the Standard Plan before deleting this workspace.',
      'Please pay your outstanding invoice before deleting this workspace.',
      'Please purchase enough tokens to clear your outstanding token balance before deleting this workspace.'
    ]);
  });

  it('counts blockers of the same kind', () => {
    const lines = instructionLines(
      [{ type: 'UNPAID_INVOICE' }, { type: 'UNPAID_INVOICE' }],
      wording
    );
    expect(lines).toEqual([
      'Please pay your 2 outstanding invoices before deleting this workspace.'
    ]);
  });

  it('falls back to plain words when the host app set no terminology', () => {
    const lines = instructionLines([{ type: 'ACTIVE_SUBSCRIPTION' }]);
    expect(lines).toEqual([
      'Downgrade your current subscription and switch to the standard plan before deleting this organization.'
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
      'Downgrade your current subscription and switch to the Standard Plan before deleting this workspace. ' +
        'Please purchase enough tokens to clear your outstanding token balance before deleting this workspace.'
    );
  });

  it('returns the generic message when the error carries no blockers', () => {
    const err = new ConnectError('blocked', Code.FailedPrecondition);
    expect(deleteBlockedDescription(err, wording)).toBe(
      GENERIC_DELETE_BLOCKED_MESSAGE
    );
  });
});
