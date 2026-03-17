import { Text, type DataTableColumnDef } from "@raystack/apsara";
import type { ServiceUser, User } from "@raystack/proton/frontier";
import { TerminologyEntity } from "../../hooks/useAdminTerminology";

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
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Email",
      accessorKey: "email",
      filterVariant: "text",
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
          <Text
            style={{ cursor: "pointer" }}
            onClick={() => onNavigateToOrg?.(org_id)}
          >
            {org_id}
          </Text>
        ) : (
          "-"
        );
      },
    },
  ];
};
