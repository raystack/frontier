import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  CheckProjectDeleteRequestSchema,
  FrontierServiceQueries
} from '@raystack/proton/frontier';
import {
  useDeleteFlow,
  type UseDeleteFlowOptions
} from '../../../hooks/useDeleteFlow';
import type { DeleteProjectPayload } from '../components/delete-project-dialog';

export interface UseProjectDeleteFlowOptions
  extends Pick<
    UseDeleteFlowOptions<DeleteProjectPayload>,
    'confirmDialogHandle' | 'blockedDialogHandle'
  > {
  menuHandle: { close: () => void };
}

// The project delete flow for a menu item. The menu stays open while the
// check runs and closes when the next dialog opens.
export function useProjectDeleteFlow({
  menuHandle,
  ...handles
}: UseProjectDeleteFlowOptions) {
  const { mutateAsync: checkProjectDelete } = useMutation(
    FrontierServiceQueries.checkProjectDelete
  );
  const { startDelete, ...flow } = useDeleteFlow({
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

  return { ...flow, startDelete: startDeleteFromMenu };
}
