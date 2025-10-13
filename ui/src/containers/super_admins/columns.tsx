import type { DataTableColumnDef } from "@raystack/apsara";
import type { ServiceUser, User } from "@raystack/proton/frontier";
import { Link } from "react-router-dom";

export const getColumns: () => DataTableColumnDef<
  User | ServiceUser,
  unknown
>[] = () => {
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
      header: "Organization",
      accessorKey: "orgId",
      cell: (info) => {
        const org_id = info.getValue() as string;
        return org_id ? (
          <Link to={`/organizations/${org_id}`}>{org_id}</Link>
        ) : (
          "-"
        );
      },
    },
  ];
};
