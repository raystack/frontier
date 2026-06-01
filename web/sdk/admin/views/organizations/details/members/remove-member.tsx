import type { SearchOrganizationUsersResponse_OrganizationUser } from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  RemoveOrganizationMemberRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation } from "@connectrpc/connect-query";
import { Button, Dialog, Flex, Text, toastManager } from "@raystack/apsara-v1";
import { handleConnectError } from "~/utils/error";
import { useTerminology } from "../../../../hooks/useTerminology";

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
  const t = useTerminology();
  const { mutateAsync: removeOrganizationMember, isPending } = useMutation(
    FrontierServiceQueries.removeOrganizationMember,
  );

  async function onSubmit() {
    try {
      if (!user) return;
      await removeOrganizationMember(
        create(RemoveOrganizationMemberRequestSchema, {
          orgId: organizationId,
          principalId: user?.id || "",
          principalType: "app/user",
        }),
      );
      if (onRemove) {
        onRemove(user);
      }
      toastManager.add({ title: `${t.member({ case: "capital" })} removed successfully`, type: "success" });
    } catch (error) {
      console.error(error);
      handleConnectError(error, {
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
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>Remove {t.member({ case: "capital" })}</Dialog.Title>
          <Dialog.CloseButton data-test-id="remove-member-close-button" />
        </Dialog.Header>
        <Dialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Removing this {t.member({ case: "lower" })} will revoke all their access to the{" "}
              {t.organization({ case: "lower" })}. This action cannot be undone. The {t.member({ case: "lower" })} will lose
              all assigned roles and permissions immediately.
            </Text>
            <Text variant="secondary">
              Are you sure you want to remove this {t.member({ case: "lower" })}?
            </Text>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close
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
            loading={isPending}
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
