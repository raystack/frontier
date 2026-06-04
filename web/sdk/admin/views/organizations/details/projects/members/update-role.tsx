import styles from "./members.module.css";
import type {
  SearchProjectUsersResponse_ProjectUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  SetProjectMemberRoleRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { ConnectError } from "@connectrpc/connect";
import { useMutation } from "@connectrpc/connect-query";
import {
  AlertDialog,
  Button,
  Text,
  toastManager,
} from "@raystack/apsara";
import { SCOPES } from "~/admin/utils/constants";

export type UpdateRolePayload = {
  user: SearchProjectUsersResponse_ProjectUser;
  role: Role;
};

interface UpdateRoleProps {
  handle: ReturnType<typeof AlertDialog.createHandle<UpdateRolePayload>>;
  projectId: string;
  onRoleUpdate: () => void;
}

export const UpdateRole = ({
  handle,
  projectId,
  onRoleUpdate,
}: UpdateRoleProps) => {
  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as UpdateRolePayload | undefined;
        return payload ? (
          <UpdateRoleContent
            payload={payload}
            projectId={projectId}
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
  projectId,
  onClose,
  onRoleUpdate,
}: {
  payload: UpdateRolePayload;
  projectId: string;
  onClose: () => void;
  onRoleUpdate: () => void;
}) {
  const { mutateAsync: setProjectMemberRole, isPending } = useMutation(
    FrontierServiceQueries.setProjectMemberRole,
  );

  async function onSubmit() {
    try {
      await setProjectMemberRole(
        create(SetProjectMemberRoleRequestSchema, {
          projectId,
          principalId: payload.user.id || "",
          principalType: SCOPES.USER,
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
    <AlertDialog.Content
      overlay={{ className: styles["action-dialog-overlay"] }}
      className={styles["action-dialog-content"]}
    >
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
