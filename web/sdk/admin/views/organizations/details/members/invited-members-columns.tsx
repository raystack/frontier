import {
  AlertDialog,
  IconButton,
  Menu,
  Text,
  type DataTableColumnDef,
} from "@raystack/apsara";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import type { Invitation } from "@raystack/proton/frontier";
import { DeleteIcon } from "~/admin/assets/icons/DeleteIcon";
import {
  formatInviteExpiry,
  formatTimestamp,
  type TimeStamp,
} from "~/admin/utils/connect-timestamp";
import type { RemoveInvitePayload } from "./remove-invite-dialog";
import styles from "./members.module.css";

interface GetColumnsOptions {
  /** Role id → title, from useOrganizationRoles. */
  roleTitleById: Map<string, string>;
  removeInviteHandle: ReturnType<
    typeof AlertDialog.createHandle<RemoveInvitePayload>
  >;
}

const seconds = (timestamp?: TimeStamp) => Number(timestamp?.seconds ?? 0);

export const getInvitedMembersColumns = ({
  roleTitleById,
  removeInviteHandle,
}: GetColumnsOptions): DataTableColumnDef<Invitation, unknown>[] => [
    {
      // Invitations carry no user record — user_id is the invited email.
      accessorKey: "userId",
      header: "Email",
      classNames: {
        header: styles["invites-email-column"],
        cell: styles["invites-email-column"],
      },
      cell: ({ getValue }) => (getValue() as string) || "-",
      enableSorting: true,
    },
    {
      accessorKey: "roleIds",
      header: "Role",
      cell: ({ getValue }) => {
        const titles = (getValue() as string[])
          .map((id) => roleTitleById.get(id))
          .filter(Boolean);
        return titles.join(", ") || "-";
      },
    },
    {
      accessorKey: "expiresAt",
      id: "status",
      header: "Status",
      cell: ({ row }) =>
        formatInviteExpiry(row.original.expiresAt).isExpired
          ? "Expired"
          : "Pending",
    },
    {
      accessorKey: "createdAt",
      header: "Invited on",
      cell: ({ row }) => formatTimestamp(row.original.createdAt),
      // Timestamps are objects, so the default comparator can't order them.
      sortingFn: (a, b) =>
        seconds(a.original.createdAt) - seconds(b.original.createdAt),
      enableSorting: true,
    },
    {
      accessorKey: "expiresAt",
      header: "Expiry",
      cell: ({ row }) => {
        const { text, isExpired } = formatInviteExpiry(row.original.expiresAt);
        return <Text variant={isExpired ? "danger" : undefined}>{text}</Text>;
      },
      sortingFn: (a, b) =>
        seconds(a.original.expiresAt) - seconds(b.original.expiresAt),
      enableSorting: true,
    },
    {
      accessorKey: "id",
      header: "",
      // Its accessor is a string, so search would match invite ids without this.
      enableGlobalFilter: false,
      classNames: {
        header: styles["invites-action-column"],
        cell: styles["invites-action-column"],
      },
      cell: ({ row }) => {
        // Offered on every row; only the confirmation copy differs.
        const { isExpired } = formatInviteExpiry(row.original.expiresAt);

        return (
          <Menu>
            <Menu.Trigger
              render={
                <IconButton
                  size={3}
                  data-test-id="admin-org-invites-action-menu">
                  <DotsHorizontalIcon />
                </IconButton>
              }
            />
            <Menu.Content>
              <Menu.Item
                leadingIcon={<DeleteIcon />}
                className={styles["invites-remove-item"]}
                onClick={() =>
                  removeInviteHandle.openWithPayload({
                    inviteId: row.original.id,
                    email: row.original.userId,
                    isExpired,
                  })
                }
                data-test-id="admin-org-invites-remove-action"
              >
                Remove
              </Menu.Item>
            </Menu.Content>
          </Menu>
        );
      },
    },
  ];
