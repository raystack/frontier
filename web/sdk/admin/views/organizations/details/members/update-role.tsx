import type {
  SearchOrganizationUsersResponse_OrganizationUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  SetOrganizationMemberRoleRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation } from "@connectrpc/connect-query";
import {
  AlertDialog,
  Button,
  Text,
  toastManager,
} from "@raystack/apsara";
import { ConnectError } from "@connectrpc/connect";

export type UpdateRolePayload = {
  user: SearchOrganizationUsersResponse_OrganizationUser;
  role: Role;
};

interface UpdateRoleProps {
  handle: ReturnType<typeof AlertDialog.createHandle<UpdateRolePayload>>;
  organizationId: string;
  onRoleUpdate: () => void;
}

export const UpdateRole = ({
  handle,
  organizationId,
  onRoleUpdate,
}: UpdateRoleProps) => {
  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as UpdateRolePayload | undefined;
        return payload ? (
          <UpdateRoleContent
            payload={payload}
            organizationId={organizationId}
            onClose={() => handle.close()}
            onRoleUpdate={onRoleUpdate}
          />
        ) : null;
      }}
    </AlertDialog>
  );
};

function UpdateRoleContent({
  payload,
  organizationId,
  onClose,
  onRoleUpdate,
}: {
  payload: UpdateRolePayload;
  organizationId: string;
  onClose: () => void;
  onRoleUpdate: () => void;
}) {
  const { mutateAsync: setMemberRole, isPending } = useMutation(
    FrontierServiceQueries.setOrganizationMemberRole,
  );

  async function onSubmit() {
    try {
      await setMemberRole(
        create(SetOrganizationMemberRoleRequestSchema, {
          orgId: organizationId,
          userId: payload.user.id,
          roleId: payload.role.id,
        }),
      );

      onRoleUpdate();
      toastManager.add({
        title: "Role assigned successfully",
        type: "success",
      });
      onClose();
    } catch (error) {
      toastManager.add({
        title: "Failed to assign role",
        description: error instanceof ConnectError ? error.message : undefined,
        type: "error",
      });
      console.error(error);
    }
  }

  return (
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Update role</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Text variant="secondary">
          This will grant additional permissions to the user based on the new
          role.
        </Text>
      </AlertDialog.Body>
      <AlertDialog.Footer>
        <Button
          type="button"
          variant="outline"
          color="neutral"
          onClick={onClose}
          disabled={isPending}
          data-test-id="assign-role-cancel-button"
        >
          Cancel
        </Button>
        <Button
          type="submit"
          data-test-id="assign-role-update-button"
          loading={isPending}
          loaderText="Updating..."
          onClick={onSubmit}
        >
          Update
        </Button>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  );
}
