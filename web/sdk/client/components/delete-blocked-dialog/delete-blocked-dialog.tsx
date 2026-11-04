'use client';

import { createContext, useContext } from 'react';
import { AlertDialog, Button, Flex, Text } from '@raystack/apsara';
import type {
  DeleteBlocker,
  DeleteBlockedContext,
  DeleteBlockedPayload
} from '../../../shared/types';
import { useFrontier } from '../../contexts/FrontierContext';
import {
  useTerminology,
  type TerminologyMap
} from '../../hooks/useTerminology';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  blockerLines,
  type BlockerWording
} from '../../utils/delete-blockers';
import styles from './delete-blocked-dialog.module.css';

export type DeleteBlockedDialogHandle = ReturnType<
  typeof AlertDialog.createHandle<DeleteBlockedPayload>
>;

function defaultContent(
  context: DeleteBlockedContext,
  t: TerminologyMap,
  messages: Record<string, string>,
  wording: BlockerWording
): { title: string; description: string; lines: string[] } {
  const { entityLabel } = context;
  switch (context.entity) {
    case 'organization':
      return {
        title: `Unable to delete ${entityLabel}`,
        description: `Please resolve the following issues before deleting this ${entityLabel}:`,
        lines: blockerLines(context.blockers, messages, wording)
      };
    case 'project':
      return {
        title: `Unable to delete ${entityLabel}`,
        description: `This ${entityLabel} cannot be deleted because it contains resources that need to be addressed first.`,
        lines: blockerLines(context.blockers, messages, wording)
      };
    case 'user': {
      const org = t.organization({ case: 'lower' });
      if (context.soleOwnerOf.length > 0) {
        return {
          title: `Transfer ${org} ownership`,
          description: `You’re the only owner of one or more ${t.organization({
            case: 'lower',
            plural: true
          })}. Before deleting your account, transfer ownership of each affected ${org} to another member.\n\nOnce ownership has been transferred, you can proceed with account deletion.`,
          lines: []
        };
      }
      const blockers = context.organizationsToDelete.flatMap(
        organization => organization.blockers
      );
      const orgs = t.organization({
        case: 'lower',
        plural: context.organizationsToDelete.length > 1
      });
      return {
        title: 'Unable to delete account',
        description: `Please resolve the following issues in your ${orgs} before deleting your account:`,
        lines:
          blockers.length > 0
            ? blockerLines(blockers, messages, wording)
            : [GENERIC_DELETE_BLOCKED_MESSAGE]
      };
    }
  }
}

const DeleteBlockedDialogContext = createContext<DeleteBlockedContext | null>(
  null
);

export interface DeleteBlockedDialogRootProps {
  handle: DeleteBlockedDialogHandle;
}

export function DeleteBlockedDialogRoot({
  handle
}: DeleteBlockedDialogRootProps) {
  return (
    <AlertDialog handle={handle}>
      {({ payload }) =>
        payload ? (
          <DeleteBlockedContent
            payload={payload}
            close={() => handle.close()}
          />
        ) : null
      }
    </AlertDialog>
  );
}

export function DeleteBlockedContent({
  payload,
  close
}: {
  payload: DeleteBlockedPayload;
  close: () => void;
}) {
  const t = useTerminology();
  const { slots } = useFrontier();
  const context: DeleteBlockedContext = {
    ...payload,
    entityLabel: t[payload.entity]({ case: 'lower' }),
    close
  };
  const body = slots.deleteBlockedDialog?.(context);
  const isEmpty = body == null || typeof body === 'boolean';
  return (
    <DeleteBlockedDialogContext.Provider value={context}>
      {isEmpty ? <DeleteBlockedDialog /> : body}
    </DeleteBlockedDialogContext.Provider>
  );
}

export interface DeleteBlockedDialogProps {
  title?: string;
  description?: string;
  blockers?: DeleteBlocker[];
  messages?: Record<string, string>;
}

export function DeleteBlockedDialog({
  title,
  description,
  blockers,
  messages = {}
}: DeleteBlockedDialogProps) {
  const context = useContext(DeleteBlockedDialogContext);
  const t = useTerminology();
  const { basePlan } = useFrontier();
  if (!context) {
    throw new Error(
      'DeleteBlockedDialog must be returned from the deleteBlockedDialog slot on FrontierProvider.'
    );
  }

  const wording = {
    organizationLabel: t.organization({ case: 'lower' }),
    basePlanTitle: basePlan?.title
  };
  const content = defaultContent(context, t, messages, wording);
  const lines = blockers
    ? blockerLines(blockers, messages, wording)
    : content.lines;

  return (
    <AlertDialog.Content className={styles.content}>
      <AlertDialog.Header>
        <AlertDialog.Title>{title ?? content.title}</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Flex direction="column" gap={5}>
          <Text size="small" variant="secondary" className={styles.description}>
            {description ?? content.description}
          </Text>
          {lines.length > 0 && (
            <Text
              size="small"
              variant="secondary"
              render={<ul className={styles.list} />}
            >
              {lines.map(line => (
                <li key={line}>{line}</li>
              ))}
            </Text>
          )}
        </Flex>
      </AlertDialog.Body>
      <AlertDialog.Footer>
        <AlertDialog.Close
          render={
            <Button
              variant="outline"
              color="neutral"
              data-test-id="frontier-sdk-delete-blocked-close-btn"
            >
              Close
            </Button>
          }
        />
      </AlertDialog.Footer>
    </AlertDialog.Content>
  );
}
