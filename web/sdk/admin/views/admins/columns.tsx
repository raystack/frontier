import { Button, type DataTableColumnDef } from "@raystack/apsara";
import type { ServiceUser, User } from "@raystack/proton/frontier";
import { TerminologyEntity } from "../../hooks/useTerminology";
import styles from "./admins.module.css";

export const getColumns: (options?: {
  onNavigateToOrg?: (orgId: string) => void;
  t?: {
    organization: TerminologyEntity;
  };
}) => DataTableColumnDef<
  User | ServiceUser,
  unknown
>[] = ({ onNavigateToOrg, t } = {}) => {
  return [
    {
      header: "Title",
      accessorKey: "title",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      filterType: "string",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Email",
      accessorKey: "email",
      filterType: "string",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Status",
      accessorKey: "state",
      meta: {
        data: [
          { label: "Enabled", value: "enabled" },
          { label: "Disabled", value: "disabled" },
        ],
      },
      cell: (info) => info.getValue(),
      footer: (props) => props.column.id,
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      header: t?.organization({ case: "capital" }) || "Organization",
      accessorKey: "orgId",
      cell: (info) => {
        const org_id = info.getValue() as string;
        return org_id ? (
          <Button
            variant="text"
            onClick={() => onNavigateToOrg?.(org_id)}
            data-test-id="frontier-admin-navigate-to-org-btn"
          >
            {org_id}
          </Button>
        ) : (
          "-"
        );
      },
    },
  ];
};
