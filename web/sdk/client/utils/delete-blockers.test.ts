import { Code, ConnectError } from '@connectrpc/connect';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  ORGANIZATION_DISABLED,
  UNKNOWN_BLOCKER,
  blockerLines,
  blockersFromError
} from './delete-blockers';

const wording = { organizationLabel: 'workspace', basePlanTitle: 'Standard Plan' };

const blocker = (type: string, subject = '') => ({ type, subject, count: 0 });

describe('blockerLines', () => {
  it('writes the SDK line for each known type', () => {
    const lines = blockerLines(
      [
        blocker('ACTIVE_SUBSCRIPTION'),
        blocker('UNPAID_INVOICE'),
        blocker('NEGATIVE_TOKEN_BALANCE'),
        blocker(ORGANIZATION_DISABLED)
      ],
      {},
      wording
    );
    expect(lines).toEqual([
      'Active subscription: Downgrade your subscription to the Standard Plan before deleting this workspace.',
      'Outstanding invoice: Pay all outstanding invoices before deleting this workspace.',
      'Negative token balance: Purchase enough tokens to clear your outstanding token balance before deleting this workspace.',
      'Disabled workspace: Contact your administrator to enable this workspace before deleting it.'
    ]);
  });

  it('falls back to plain words when the host app has no base plan', () => {
    expect(
      blockerLines([blocker('ACTIVE_SUBSCRIPTION')], {}, {
        organizationLabel: 'organization'
      })
    ).toEqual([
      'Active subscription: Downgrade your subscription to the standard plan before deleting this organization.'
    ]);
  });

  it('prefers a message for the subject, then for the type', () => {
    const lines = blockerLines(
      [
        blocker('RESOURCES_EXIST', 'app/thing'),
        blocker('RESOURCES_EXIST', 'app/other'),
        blocker('UNPAID_INVOICE', 'invoice-1')
      ],
      {
        'app/thing': 'Things exist.',
        RESOURCES_EXIST: 'Resources exist.',
        UNPAID_INVOICE: 'Pay up.'
      },
      wording
    );
    expect(lines).toEqual(['Things exist.', 'Resources exist.', 'Pay up.']);
  });

  it('shows identical lines once', () => {
    const lines = blockerLines(
      [blocker('UNPAID_INVOICE', 'invoice-1'), blocker('UNPAID_INVOICE', 'invoice-2')],
      {},
      wording
    );
    expect(lines).toHaveLength(1);
  });

  it('adds the generic line once, last, for blockers with no line', () => {
    const lines = blockerLines(
      [
        blocker('RESOURCES_EXIST', 'app/thing'),
        blocker('UNPAID_INVOICE'),
        blocker('SOMETHING_NEW')
      ],
      {},
      wording
    );
    expect(lines).toEqual([
      'Outstanding invoice: Pay all outstanding invoices before deleting this workspace.',
      GENERIC_DELETE_BLOCKED_MESSAGE
    ]);
  });

  it('gives no lines for no blockers', () => {
    expect(blockerLines([], {}, wording)).toEqual([]);
  });
});

describe('blockersFromError', () => {
  it('reads the blockers out of the precondition failure detail', () => {
    const err = new ConnectError('blocked', Code.FailedPrecondition);
    err.details.push({
      type: 'google.rpc.PreconditionFailure',
      value: new Uint8Array(),
      debug: {
        violations: [
          { type: 'ACTIVE_SUBSCRIPTION', subject: 'sub-1', description: 'developer text' },
          { type: 'RESOURCES_EXIST', subject: 'app/thing' },
          { subject: 'no-type' }
        ]
      }
    });
    expect(blockersFromError(err)).toEqual([
      { type: 'ACTIVE_SUBSCRIPTION', subject: 'sub-1', count: 0 },
      { type: 'RESOURCES_EXIST', subject: 'app/thing', count: 0 }
    ]);
  });

  it('treats an error with no precondition failure as a disabled organization', () => {
    const err = new ConnectError('org is disabled', Code.FailedPrecondition);
    expect(blockersFromError(err)).toEqual([
      { type: ORGANIZATION_DISABLED, subject: '', count: 0 }
    ]);
  });

  it('gives an unknown blocker when the violations cannot be read', () => {
    const err = new ConnectError('blocked', Code.FailedPrecondition);
    err.details.push({
      type: 'google.rpc.PreconditionFailure',
      value: new Uint8Array()
    });
    const blockers = blockersFromError(err);
    expect(blockers).toEqual([{ type: UNKNOWN_BLOCKER, subject: '', count: 0 }]);
    expect(blockerLines(blockers, {}, wording)).toEqual([
      GENERIC_DELETE_BLOCKED_MESSAGE
    ]);
  });
});
