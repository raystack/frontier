'use client';

import { useEffect, useState } from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  Checkbox,
  Field,
  Flex,
  Input,
  Skeleton,
  Text,
  toastManager
} from '@raystack/apsara';
import type { Slot } from '../../../shared/types';
import { useFrontier } from '../../contexts/FrontierContext';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  toUserDeleteBlocked
} from '../../utils/delete-blockers';
import { DeleteBlockedContent } from '../delete-blocked-dialog';
import { handleConnectError } from '~/utils/error';
import styles from './delete-user-dialog.module.css';

const TITLE = 'Delete Account';

const showErrorToast = () =>
  toastManager.add({
    title: 'Something went wrong',
    description: 'Please try again later or contact support.',
    type: 'error'
  });

export type DeleteUserDialogHandle = ReturnType<
  typeof AlertDialog.createHandle<null>
>;

const createDeleteUserDialogHandle = (): DeleteUserDialogHandle =>
  AlertDialog.createHandle<null>();

export interface DeleteUserDialogProps {
  /** A trigger that stays mounted while the dialog is open. */
  trigger?: Slot<{ onClick: () => void }>;
  /** Opens the dialog from elsewhere, such as a menu that unmounts when it closes. */
  handle?: DeleteUserDialogHandle;
  /** Called after the account is deleted. The host app decides where to go. */
  onDeleted: () => void;
}

/**
 * The whole account delete flow. It opens first and checks inside, so any
 * trigger can open it. It never shows the confirmation while the delete is
 * blocked.
 */
export function DeleteUserDialog({
  trigger,
  handle,
  onDeleted
}: DeleteUserDialogProps) {
  const [ownHandle] = useState(createDeleteUserDialogHandle);
  const dialogHandle = handle ?? ownHandle;
  return (
    <>
      {trigger?.({ onClick: () => dialogHandle.open(null) })}
      <AlertDialog handle={dialogHandle}>
        <AlertDialog.Content className={styles.content}>
          <DeleteUserFlow handle={dialogHandle} onDeleted={onDeleted} />
        </AlertDialog.Content>
      </AlertDialog>
    </>
  );
}

DeleteUserDialog.createHandle = createDeleteUserDialogHandle;

interface DeleteUserFlowProps {
  handle: DeleteUserDialogHandle;
  onDeleted: () => void;
}

// Mounts only while the dialog is open, so every open runs a new check.
function DeleteUserFlow({ handle, onDeleted }: DeleteUserFlowProps) {
  const { data: check, error, refetch } = useQuery(
    FrontierServiceQueries.checkCurrentUserDelete,
    {},
    { retry: false, gcTime: 0, refetchOnWindowFocus: false }
  );

  useEffect(() => {
    if (error) {
      handle.close();
      showErrorToast();
    }
  }, [error, handle]);

  // A refused delete has no grouping by organization, so check again. A
  // failed check closes the dialog through the error effect above.
  async function onDeleteRefused() {
    const { data, isSuccess } = await refetch();
    if (isSuccess && data.canDelete) {
      toastManager.add({
        title: 'Unable to delete account',
        description: GENERIC_DELETE_BLOCKED_MESSAGE,
        type: 'error'
      });
    }
  }

  if (!check) {
    return <DeleteUserChecking />;
  }
  if (!check.canDelete) {
    return (
      <DeleteBlockedContent
        payload={toUserDeleteBlocked(check)}
        close={() => handle.close()}
      />
    );
  }
  return (
    <DeleteUserForm
      onDeleted={() => {
        handle.close();
        onDeleted();
      }}
      onDeleteRefused={onDeleteRefused}
    />
  );
}

function DeleteUserChecking() {
  return (
    <>
      <AlertDialog.Header>
        <AlertDialog.Title>{TITLE}</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Flex direction="column" gap={3}>
          <Skeleton height="16px" />
          <Skeleton height="16px" width="60%" />
        </Flex>
      </AlertDialog.Body>
      <AlertDialog.Footer>
        <DialogCancel />
      </AlertDialog.Footer>
    </>
  );
}

interface DeleteUserFormProps {
  onDeleted: () => void;
  onDeleteRefused: () => void;
}

function DeleteUserForm({ onDeleted, onDeleteRefused }: DeleteUserFormProps) {
  const { user } = useFrontier();
  const [email, setEmail] = useState('');
  const [isAcknowledged, setIsAcknowledged] = useState(false);
  const { mutateAsync: deleteCurrentUser, isPending } = useMutation(
    FrontierServiceQueries.deleteCurrentUser
  );
  const isEmailConfirmed =
    !!user?.email && email.trim().toLowerCase() === user.email.toLowerCase();

  async function onDelete() {
    try {
      await deleteCurrentUser({});
      toastManager.add({ title: 'Account deleted', type: 'success' });
      onDeleted();
    } catch (error) {
      handleConnectError(error, {
        FailedPrecondition: onDeleteRefused,
        Default: showErrorToast
      });
    }
  }

  return (
    <>
      <AlertDialog.Header>
        <AlertDialog.Title>{TITLE}</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Flex direction="column" gap={6}>
          <Text size="small" variant="secondary">
            This action is irreversible. Deleting your account will permanently
            remove your account and associated data. Do you want to continue?
          </Text>
          <Field label="Please type your email id.">
            <Input
              size="large"
              type="email"
              value={email}
              onChange={event => setEmail(event.target.value)}
              data-test-id="frontier-sdk-delete-account-email-input"
            />
          </Field>
          <Flex gap={3} align="start">
            <Checkbox
              checked={isAcknowledged}
              onCheckedChange={(checked: boolean) => setIsAcknowledged(checked)}
              data-test-id="frontier-sdk-delete-account-checkbox"
            />
            <Text size="small" variant="secondary">
              I acknowledge that I understand my account and associated data
              will be permanently deleted and want to proceed
            </Text>
          </Flex>
        </Flex>
      </AlertDialog.Body>
      <AlertDialog.Footer>
        <DialogCancel disabled={isPending} />
        <Button
          variant="solid"
          color="danger"
          onClick={onDelete}
          disabled={!isEmailConfirmed || !isAcknowledged || isPending}
          loading={isPending}
          loaderText="Deleting..."
          data-test-id="frontier-sdk-delete-account-btn"
        >
          Delete
        </Button>
      </AlertDialog.Footer>
    </>
  );
}

function DialogCancel({ disabled }: { disabled?: boolean }) {
  return (
    <AlertDialog.Close
      render={
        <Button
          variant="outline"
          color="neutral"
          disabled={disabled}
          data-test-id="frontier-sdk-delete-account-cancel-btn"
        >
          Cancel
        </Button>
      }
    />
  );
}
