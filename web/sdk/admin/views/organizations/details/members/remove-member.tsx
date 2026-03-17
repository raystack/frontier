import type { SearchOrganizationUsersResponse_OrganizationUser } from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  RemoveOrganizationUserRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation } from "@connectrpc/connect-query";
import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara";
import { ConnectError } from "@connectrpc/connect";
import { useAdminTerminology } from "../../../../hooks/useAdminTerminology";

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
  const t = useAdminTerminology();
  const { mutateAsync: removeOrganizationUser, isPending } = useMutation(
    FrontierServiceQueries.removeOrganizationUser,
  );

  async function onSubmit() {
    try {
      if (!user) return;
      await removeOrganizationUser(
        create(RemoveOrganizationUserRequestSchema, {
          id: organizationId,
          userId: user?.id || "",
        }),
      );
      if (onRemove) {
        onRemove(user);
      }
      toast.success(`${t.member({ case: "capital" })} removed successfully`);
    } catch (error) {
      const message =
        error instanceof ConnectError
          ? error.message
          : "Unknown error";
      toast.error(`Failed to remove ${t.member({ case: "lower" })}: ${message}`);
      console.error(error);
    }
  }

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content width={400}>
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
