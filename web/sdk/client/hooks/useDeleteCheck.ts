import { useState } from 'react';
import { AlertDialog, toastManager } from '@raystack/apsara';
import type { Blocker } from '@raystack/proton/frontier';
import type { DeleteBlocker, DeleteBlockedPayload } from '../../shared/types';
import type { DeleteBlockedDialogHandle } from '../components/delete-blocked-dialog';
import { blockersFromError, toDeleteBlockers } from '../utils/delete-blockers';
import { useTerminology } from './useTerminology';
import { handleConnectError } from '~/utils/error';

export interface UseDeleteCheckOptions<TPayload> {
  entity: DeleteBlockedPayload['entity'];
  check: (
    payload: TPayload
  ) => Promise<{ canDelete: boolean; blockers: Blocker[] }>;
  confirmDialogHandle: ReturnType<typeof AlertDialog.createHandle<TPayload>>;
  blockedDialogHandle: DeleteBlockedDialogHandle;
}

export function useDeleteCheck<TPayload>({
  entity,
  check,
  confirmDialogHandle,
  blockedDialogHandle
}: UseDeleteCheckOptions<TPayload>) {
  const t = useTerminology();
  const [isChecking, setIsChecking] = useState(false);

  function showBlocked(blockers: DeleteBlocker[]) {
    confirmDialogHandle.close();
    blockedDialogHandle.openWithPayload({ entity, blockers });
  }

  function handleDeleteError(error: unknown) {
    handleConnectError(error, {
      FailedPrecondition: err => showBlocked(blockersFromError(err)),
      PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
      NotFound: () => toastManager.add({ title: 'Not found', description: `This ${t[entity]({ case: 'lower' })} no longer exists.`, type: 'error' }),
      Default: () => toastManager.add({ title: 'Something went wrong', description: 'Please try again later or contact support.', type: 'error' }),
    });
  }

  async function startDelete(payload: TPayload) {
    setIsChecking(true);
    try {
      const { canDelete, blockers } = await check(payload);
      if (canDelete) {
        confirmDialogHandle.openWithPayload(payload);
      } else {
        showBlocked(toDeleteBlockers(blockers));
      }
    } catch (error) {
      handleDeleteError(error);
    } finally {
      setIsChecking(false);
    }
  }

  return { startDelete, isChecking, handleDeleteError };
}
