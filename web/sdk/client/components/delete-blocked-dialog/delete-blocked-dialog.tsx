'use client';

import { createContext, useContext } from 'react';
import { AlertDialog, Button, Flex, Text } from '@raystack/apsara';
import type {
  DeleteBlocker,
  DeleteBlockedContext,
  DeleteBlockedPayload
} from '../../../shared/types';
import { useFrontier } from '../../contexts/FrontierContext';
import { useTerminology } from '../../hooks/useTerminology';
import { blockerLines } from '../../utils/delete-blockers';
import styles from './delete-blocked-dialog.module.css';

export type DeleteBlockedDialogHandle = ReturnType<
  typeof AlertDialog.createHandle<DeleteBlockedPayload>
>;

const DEFAULT_COPY: Record<
  DeleteBlockedPayload['entity'],
  (entityLabel: string) => { title: string; description: string }
> = {
  organization: entityLabel => ({
    title: `Unable to delete ${entityLabel}`,
    description: `Please resolve the following issues before deleting this ${entityLabel}:`
  }),
  project: entityLabel => ({
    title: `Unable to delete ${entityLabel}`,
    description: `This ${entityLabel} cannot be deleted because it contains resources that need to be addressed first.`
  })
};

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

function DeleteBlockedContent({
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

  const copy = DEFAULT_COPY[context.entity](context.entityLabel);
  const lines = blockerLines(blockers ?? context.blockers, messages, {
    organizationLabel: t.organization({ case: 'lower' }),
    basePlanTitle: basePlan?.title
  });

  return (
    <AlertDialog.Content className={styles.content}>
      <AlertDialog.Header>
        <AlertDialog.Title>{title ?? copy.title}</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Flex direction="column" gap={5}>
          <Text size="small" variant="secondary">
            {description ?? copy.description}
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
