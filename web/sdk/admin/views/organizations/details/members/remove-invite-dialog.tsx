import { create } from "@bufbuild/protobuf";
import {
  useMutation,
  createConnectQueryKey,
  useTransport,
} from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import {
  FrontierServiceQueries,
  DeleteOrganizationInvitationRequestSchema,
} from "@raystack/proton/frontier";
import { AlertDialog, Button, toastManager } from "@raystack/apsara";
import { handleConnectError } from "~/utils/error";

export type RemoveInvitePayload = {
  inviteId: string;
  email: string;
  isExpired: boolean;
};

const DESCRIPTION = {
  pending: (email: string) =>
    `The invitation for ${email} will be revoked and the link in their email will stop working.`,
  expired: (email: string) =>
    `The invitation sent to ${email} has already expired. Removing it only clears the entry from this list.`,
};

interface RemoveInviteDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<RemoveInvitePayload>>;
  organizationId: string;
}

export const RemoveInviteDialog = ({
  handle,
  organizationId,
}: RemoveInviteDialogProps) => {
  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as RemoveInvitePayload | undefined;
        return payload ? (
          <RemoveInviteContent
            payload={payload}
            organizationId={organizationId}
            onClose={() => handle.close()}
          />
        ) : null;
      }}
    </AlertDialog>
  );
};

function RemoveInviteContent({
  payload,
  organizationId,
  onClose,
}: {
  payload: RemoveInvitePayload;
  organizationId: string;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutateAsync: deleteInvitation, isPending } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
  );

  async function onRemove() {
    try {
      await deleteInvitation(
        create(DeleteOrganizationInvitationRequestSchema, {
          orgId: organizationId,
          id: payload.inviteId,
        }),
      );
    } catch (error) {
      console.error(error);
      handleConnectError(error, {
        NotFound: () =>
          toastManager.add({
            title: "Invitation no longer exists",
            type: "error",
          }),
        PermissionDenied: () =>
          toastManager.add({
            title: "You don't have permission to perform this action",
            type: "error",
          }),
        Default: err =>
          toastManager.add({
            title: "Something went wrong",
            description: err.rawMessage,
            type: "error",
          }),
      });
      return;
    }

    toastManager.add({ title: "Invitation removed", type: "success" });
    onClose();

    // Outside the try: a failing refetch mustn't read as a failed delete.
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: FrontierServiceQueries.listOrganizationInvitations,
        transport,
        input: { orgId: organizationId },
        cardinality: "finite",
      }),
    });
  }

  return (
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Remove invitation</AlertDialog.Title>
        <AlertDialog.Description>
          {payload.isExpired
            ? DESCRIPTION.expired(payload.email)
            : DESCRIPTION.pending(payload.email)}
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer>
        <AlertDialog.Close
          render={
            <Button
              type="button"
              variant="outline"
              color="neutral"
              disabled={isPending}
              data-test-id="admin-remove-invite-cancel"
            >
              Cancel
            </Button>
          }
        />
        <Button
          type="button"
          variant="solid"
          color="danger"
          onClick={onRemove}
          loading={isPending}
          loaderText="Removing..."
          data-test-id="admin-remove-invite-confirm"
        >
          Remove
        </Button>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  );
}
