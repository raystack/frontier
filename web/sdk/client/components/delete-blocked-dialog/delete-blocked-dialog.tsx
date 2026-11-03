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

export type { DeleteBlockedPayload };

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
  })
};

const OpenDeleteBlockedContext = createContext<DeleteBlockedContext | null>(
  null
);

export interface DeleteBlockedDialogHostProps {
  handle: DeleteBlockedDialogHandle;
}

// The dialog the SDK opens when a delete is blocked. Its body comes from the
// deleteBlockedDialog slot on FrontierProvider, or DeleteBlockedDialog.
export function DeleteBlockedDialogHost({ handle }: DeleteBlockedDialogHostProps) {
  return (
    <AlertDialog handle={handle}>
      {({ payload }) =>
        payload ? (
          <AlertDialog.Content className={styles.content}>
            <DeleteBlockedBody
              payload={payload as DeleteBlockedPayload}
              close={() => handle.close()}
            />
          </AlertDialog.Content>
        ) : null
      }
    </AlertDialog>
  );
}

// Calls the slot inside a component of its own, so the slot can use hooks.
function DeleteBlockedBody({
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
  // A slot that renders nothing would leave an empty modal with no Close.
  const isEmpty = body == null || typeof body === 'boolean';
  return (
    <OpenDeleteBlockedContext.Provider value={context}>
      {isEmpty ? <DeleteBlockedDialog /> : body}
    </OpenDeleteBlockedContext.Provider>
  );
}

export interface DeleteBlockedDialogProps {
  /** Default: the SDK title for the entity. */
  title?: string;
  /** Default: the SDK intro for the entity. */
  description?: string;
  /** Default: the open dialog's blockers. Pass `[]` for no bullets. */
  blockers?: DeleteBlocker[];
  /** One line per resource namespace or blocker type. Missing keys use the SDK line. */
  messages?: Record<string, string>;
}

/**
 * The body of the dialog that says why a delete is blocked. Return it from
 * the `deleteBlockedDialog` slot on FrontierProvider. It renders only inside
 * the dialog the SDK has open.
 */
export function DeleteBlockedDialog({
  title,
  description,
  blockers,
  messages = {}
}: DeleteBlockedDialogProps) {
  const context = useContext(OpenDeleteBlockedContext);
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
    <>
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
    </>
  );
}
