'use client';

import { useCallback, useEffect, useState } from 'react';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  Checkbox,
  Field,
  Flex,
  Input,
  Text,
  toastManager
} from '@raystack/apsara';
import type { DeleteBlockedPayload, Slot } from '../../../shared/types';
import { useFrontier } from '../../contexts/FrontierContext';
import {
  GENERIC_DELETE_BLOCKED_MESSAGE,
  toUserDeleteBlocked
} from '../../utils/delete-blockers';
import { DeleteBlockedContent } from '../delete-blocked-dialog';
import { handleConnectError } from '~/utils/error';
import styles from './delete-user-dialog.module.css';

const showErrorToast = () =>
  toastManager.add({
    title: 'Something went wrong',
    description: 'Please try again later or contact support.',
    type: 'error'
  });

type DeleteUserPayload =
  | { step: 'confirm' }
  | {
      step: 'blocked';
      blocked: Extract<DeleteBlockedPayload, { entity: 'user' }>;
    };

export interface DeleteUserDialogHandle {
  open: () => void;
}

const deleteStarters = new WeakMap<DeleteUserDialogHandle, () => void>();

const createDeleteUserDialogHandle = (): DeleteUserDialogHandle => {
  const handle: DeleteUserDialogHandle = {
    open: () => deleteStarters.get(handle)?.()
  };
  return handle;
};

export interface DeleteUserDialogProps {
  trigger?: Slot<{ onClick: () => void; loading: boolean }>;
  handle?: DeleteUserDialogHandle;
  onDeleted: () => void;
}

export function DeleteUserDialog({
  trigger,
  handle,
  onDeleted
}: DeleteUserDialogProps) {
  const [dialogHandle] = useState(() =>
    AlertDialog.createHandle<DeleteUserPayload>()
  );
  const { mutateAsync: checkCurrentUserDelete, isPending: isChecking } =
    useMutation(FrontierServiceQueries.checkCurrentUserDelete);

  const checkAndOpen = useCallback(async () => {
    try {
      const check = await checkCurrentUserDelete({});
      dialogHandle.openWithPayload(
        check.canDelete
          ? { step: 'confirm' }
          : { step: 'blocked', blocked: toUserDeleteBlocked(check) }
      );
      return check.canDelete;
    } catch {
      dialogHandle.close();
      showErrorToast();
      return false;
    }
  }, [checkCurrentUserDelete, dialogHandle]);

  const startDelete = useCallback(() => {
    void checkAndOpen();
  }, [checkAndOpen]);

  useEffect(() => {
    if (!handle) return;
    deleteStarters.set(handle, startDelete);
    return () => {
      deleteStarters.delete(handle);
    };
  }, [handle, startDelete]);

  async function onDeleteRefused() {
    if (await checkAndOpen()) {
      toastManager.add({
        title: 'Unable to delete account',
        description: GENERIC_DELETE_BLOCKED_MESSAGE,
        type: 'error'
      });
    }
  }

  return (
    <>
      {trigger?.({ onClick: startDelete, loading: isChecking })}
      <AlertDialog handle={dialogHandle}>
        {({ payload }) => {
          if (!payload) return null;
          return payload.step === 'blocked' ? (
            <DeleteBlockedContent
              payload={payload.blocked}
              close={() => dialogHandle.close()}
            />
          ) : (
            <AlertDialog.Content className={styles.content}>
              <DeleteUserForm
                onDeleted={() => {
                  dialogHandle.close();
                  onDeleted();
                }}
                onDeleteRefused={onDeleteRefused}
              />
            </AlertDialog.Content>
          );
        }}
      </AlertDialog>
    </>
  );
}

DeleteUserDialog.createHandle = createDeleteUserDialogHandle;

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
        <AlertDialog.Title>Delete Account</AlertDialog.Title>
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
          <Flex gap={3} align="start" render={<label />}>
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
        <AlertDialog.Close
          render={
            <Button
              variant="outline"
              color="neutral"
              disabled={isPending}
              data-test-id="frontier-sdk-delete-account-cancel-btn"
            >
              Cancel
            </Button>
          }
        />
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
