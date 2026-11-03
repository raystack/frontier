import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckProjectDeleteRequestSchema,
  FrontierServiceQueries
} from '@raystack/proton/frontier';
import {
  useDeleteCheck,
  type UseDeleteCheckOptions
} from '../../../hooks/useDeleteCheck';
import type { DeleteProjectPayload } from '../components/delete-project-dialog';

export interface UseProjectDeleteCheckOptions
  extends Pick<
    UseDeleteCheckOptions<DeleteProjectPayload>,
    'confirmDialogHandle' | 'blockedDialogHandle'
  > {
  menuHandle: { close: () => void };
}

export function useProjectDeleteCheck({
  menuHandle,
  ...handles
}: UseProjectDeleteCheckOptions) {
  const { mutateAsync: checkProjectDelete } = useMutation(
    FrontierServiceQueries.checkProjectDelete
  );
  const { startDelete, ...deleteCheck } = useDeleteCheck({
    entity: 'project',
    check: ({ projectId }) =>
      checkProjectDelete(
        create(CheckProjectDeleteRequestSchema, { id: projectId })
      ),
    ...handles
  });

  async function startDeleteFromMenu(payload: DeleteProjectPayload) {
    await startDelete(payload);
    menuHandle.close();
  }

  return { ...deleteCheck, startDelete: startDeleteFromMenu };
}
