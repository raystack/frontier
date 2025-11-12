import { useState } from "react";
import type { SearchOrganizationUsersResponse_OrganizationUser } from "@raystack/proton/frontier";

import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara";
import { clients } from "~/connect/clients";
import { ConnectError } from "@connectrpc/connect";

interface RemoveMemberProps {
  organizationId: string;
  user?: SearchOrganizationUsersResponse_OrganizationUser;
  onRemove: (user: SearchOrganizationUsersResponse_OrganizationUser) => void;
  onClose: () => void;
}

export const RemoveMember = ({
  organizationId,
  user,
  onRemove,
  onClose,
}: RemoveMemberProps) => {
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function onSubmit() {
    try {
      if (!user) return;
      setIsSubmitting(true);
      const client = clients.frontier();
      await client.removeOrganizationUser({
        id: organizationId,
        userId: user?.id || "",
      });
      if (onRemove) {
        onRemove(user);
      }
      toast.success("Member removed successfully");
    } catch (error) {
      const message =
        error instanceof ConnectError
          ? error.message
          : "Unknown error";
      toast.error(`Failed to remove member: ${message}`);
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content width={400}>
        <Dialog.Header>
          <Dialog.Title>Remove Member</Dialog.Title>
          <Dialog.CloseButton data-test-id="remove-member-close-button" />
        </Dialog.Header>
        <Dialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Removing this member will revoke all their access to the
              organization. This action cannot be undone. The member will lose
              all assigned roles and permissions immediately.
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
