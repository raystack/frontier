import { useMemo, useState } from "react";
import {
  AlertDialog,
  Button,
  DataTable,
  Dialog,
  EmptyState,
  Flex,
  type DataTableSort,
} from "@raystack/apsara";
import type { Invitation } from "@raystack/proton/frontier";
import { UsersIcon } from "~/admin/assets/icons/UsersIcon";
import { useOrganizationRoles } from "~/admin/hooks/useOrganizationRoles";
import { useTerminology } from "~/admin/hooks/useTerminology";
import { InviteUsersDialog } from "../layout/invite-users-dialog";
import { getInvitedMembersColumns } from "./invited-members-columns";
import {
  RemoveInviteDialog,
  type RemoveInvitePayload,
} from "./remove-invite-dialog";
import styles from "./members.module.css";

const removeInviteHandle = AlertDialog.createHandle<RemoveInvitePayload>();

const DEFAULT_SORT: DataTableSort = { name: "createdAt", order: "desc" };

interface InvitedMembersDialogProps {
  organizationId: string;
  invitations: Invitation[];
  isLoading: boolean;
  onClose: () => void;
}

const NoInvites = () => (
  <EmptyState
    classNames={{
      container: styles["empty-state"],
      subHeading: styles["empty-state-subheading"],
    }}
    heading="No invites found"
    subHeading="Invitations that have been sent but not yet accepted will appear here."
    icon={<UsersIcon />}
  />
);

export const InvitedMembersDialog = ({
  organizationId,
  invitations,
  isLoading,
  onClose,
}: InvitedMembersDialogProps) => {
  const t = useTerminology();
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const { titleById } = useOrganizationRoles(organizationId);

  const columns = useMemo(
    () =>
      getInvitedMembersColumns({
        roleTitleById: titleById,
        removeInviteHandle,
      }),
    [titleById],
  );

  return (
    <>
      {isInviteDialogOpen ? (
        <InviteUsersDialog onOpenChange={setIsInviteDialogOpen} />
      ) : null}
      <RemoveInviteDialog
        handle={removeInviteHandle}
        organizationId={organizationId}
      />
      <Dialog open onOpenChange={onClose}>
        <Dialog.Content className={styles["invites-dialog-content"]}>
          <Dialog.Header>
            <Dialog.Title>Invited {t.member({ plural: true })}</Dialog.Title>
            <Dialog.CloseButton data-test-id="admin-org-invites-close" />
          </Dialog.Header>
          <Dialog.Body className={styles["invites-dialog-body"]}>
            {/* Client mode: the list API takes no query. */}
            <DataTable
              columns={columns}
              data={invitations}
              isLoading={isLoading}
              mode="client"
              defaultSort={DEFAULT_SORT}
            >
              <Flex
                direction="column"
                gap={5}
                className={styles["invites-table-wrapper"]}
              >
                <Flex justify="between" align="center" gap={4}>
                  {/* Sized via `width`; className lands on the inner input. */}
                  <DataTable.Search
                    placeholder="Search by email"
                    showClearButton={true}
                    width="33.33%"
                  />
                  <Button
                    onClick={() => setIsInviteDialogOpen(true)}
                    data-test-id="admin-org-invites-invite-member"
                  >
                    Invite {t.member({ case: "lower" })}
                  </Button>
                </Flex>
                <DataTable.Content
                  emptyState={<NoInvites />}
                  classNames={{
                    root: styles["invites-table-scroll"],
                    table: styles["table"],
                    header: styles["table-header"],
                  }}
                />
              </Flex>
            </DataTable>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
