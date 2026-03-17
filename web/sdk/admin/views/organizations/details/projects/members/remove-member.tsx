import { useState } from "react";
import {
  FrontierService,
  FrontierServiceQueries,
  ListPoliciesRequestSchema,
  DeletePolicyRequestSchema,
  type SearchProjectUsersResponse_ProjectUser,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation, useTransport } from "@connectrpc/connect-query";
import { createClient } from "@connectrpc/connect";
import styles from "./members.module.css";

import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara";
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
  const transport = useTransport();

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy,
  );

  async function onSubmit() {
    try {
      if (!user) return;
      setIsSubmitting(true);
      const client = createClient(FrontierService, transport);
      const policiesResp = await client.listPolicies(
        create(ListPoliciesRequestSchema, {
          projectId: projectId,
          userId: user?.id,
        }),
      );
      const policies = policiesResp.policies || [];
      await Promise.all(
        policies.map((policy) =>
          deletePolicy(
            create(DeletePolicyRequestSchema, { id: policy.id || "" }),
          ),
        ),
      );

      if (onRemove) {
        onRemove(user);
      }

      toast.success(`${t.member({ case: "capital" })} removed successfully`);
    } catch (error) {
      toast.error(`Failed to remove ${t.member({ case: "lower" })}`);
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
          <Dialog.Title>Remove {t.member({ case: "capital" })}</Dialog.Title>
          <Dialog.CloseButton data-test-id="remove-member-close-button" />
        </Dialog.Header>
        <Dialog.Body>
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
