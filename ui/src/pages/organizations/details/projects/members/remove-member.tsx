import { useState } from "react";
import { SearchProjectUsersResponseProjectUser } from "~/api/frontier";
import styles from "./members.module.css";

import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara/v1";
import { api } from "~/api";

interface RemoveMemberProps {
  projectId: string;
  user?: SearchProjectUsersResponseProjectUser;
  onRemove: (user: SearchProjectUsersResponseProjectUser) => void;
  onClose: () => void;
}

export const RemoveMember = ({
  projectId,
  user,
  onRemove,
  onClose,
}: RemoveMemberProps) => {
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function onSubmit() {
    try {
      if (!user) return;
      setIsSubmitting(true);
      const policiesResp = await api?.frontierServiceListPolicies({
        project_id: projectId,
        user_id: user?.id,
      });
      const policies = policiesResp?.data?.policies || [];
      await Promise.all(
        policies.map((policy) =>
          api?.frontierServiceDeletePolicy(policy.id as string),
        ),
      );

      if (onRemove) {
        onRemove(user);
      }

      toast.success("Member removed successfully");
    } catch (error) {
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content
        width={400}
        overlayClassName={styles["action-dialog-overlay"]}
        className={styles["action-dialog-content"]}
      >
        <Dialog.Header>
          <Dialog.Title>Remove Member</Dialog.Title>
          <Dialog.CloseButton data-test-id="remove-member-close-button" />
        </Dialog.Header>
        <Dialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Removing this member will revoke all their access to the project.
              This action cannot be undone. The member will lose all assigned
              roles and permissions immediately.
            </Text>
            <Text variant="secondary">
              Are you sure you want to remove this member?
            </Text>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              type="button"
              variant="outline"
              color="neutral"
              data-test-id="remove-member-cancel-button"
            >
              Cancel
            </Button>
          </Dialog.Close>
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
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
