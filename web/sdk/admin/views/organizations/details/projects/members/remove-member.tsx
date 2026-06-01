import { useState } from "react";
import {
  FrontierServiceQueries,
  RemoveProjectMemberRequestSchema,
  type SearchProjectUsersResponse_ProjectUser,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation } from "@connectrpc/connect-query";
import styles from "./members.module.css";

import { AlertDialog, Button, Flex, Text, toastManager } from "@raystack/apsara-v1";
import { handleConnectError } from "~/utils/error";
import { useTerminology } from "../../../../../hooks/useTerminology";

interface RemoveMemberProps {
  projectId: string;
  user?: SearchProjectUsersResponse_ProjectUser;
  onRemove: (user: SearchProjectUsersResponse_ProjectUser) => void;
  onClose: () => void;
}

export const RemoveMember = ({
  projectId,
  user,
  onRemove,
  onClose,
}: RemoveMemberProps) => {
  const t = useTerminology();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { mutateAsync: removeProjectMember } = useMutation(
    FrontierServiceQueries.removeProjectMember,
  );

  async function onSubmit() {
    try {
      if (!user) return;
      setIsSubmitting(true);
      await removeProjectMember(
        create(RemoveProjectMemberRequestSchema, {
          projectId: projectId,
          principalId: user.id,
          principalType: "app/user",
        }),
      );

      if (onRemove) {
        onRemove(user);
      }

      toastManager.add({ title: `${t.member({ case: "capital" })} removed successfully`, type: "success" });
    } catch (error) {
      handleConnectError(error, {
        NotFound: (err) =>
          toastManager.add({
            title: "Not found",
            description: err.rawMessage,
            type: "error",
          }),
        PermissionDenied: () =>
          toastManager.add({
            title: "You don't have permission to perform this action",
            type: "error",
          }),
        Default: (err) =>
          toastManager.add({
            title: `Failed to remove ${t.member({ case: "lower" })}`,
            description: err.rawMessage,
            type: "error",
          }),
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <AlertDialog open onOpenChange={onClose}>
      <AlertDialog.Content
        overlay={{ className: styles["action-dialog-overlay"] }}
        className={styles["action-dialog-content"]}
      >
        <AlertDialog.Header>
          <AlertDialog.Title>Remove {t.member({ case: "capital" })}</AlertDialog.Title>
        </AlertDialog.Header>
        <AlertDialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Removing this {t.member({ case: "lower" })} will revoke all their access to the {t.project({ case: "lower" })}.
              This action cannot be undone. The {t.member({ case: "lower" })} will lose all assigned
              roles and permissions immediately.
            </Text>
            <Text variant="secondary">
              Are you sure you want to remove this {t.member({ case: "lower" })}?
            </Text>
          </Flex>
        </AlertDialog.Body>
        <AlertDialog.Footer>
          <AlertDialog.Close
            render={
              <Button
                type="button"
                variant="outline"
                color="neutral"
                data-test-id="remove-member-cancel-button"
              >
                Cancel
              </Button>
            }
          />
          <Button
            type="submit"
            data-test-id="remove-member-submit-button"
            color="danger"
            loading={isSubmitting}
            loaderText="Removing..."
            onClick={onSubmit}
          >
            Remove
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
