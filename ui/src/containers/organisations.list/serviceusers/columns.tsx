import { CheckCircledIcon } from "@radix-ui/react-icons";
import { V1Beta1ServiceUser, V1Beta1User } from "@raystack/frontier";
import type { ColumnDef } from "@tanstack/react-table";
import { createColumnHelper } from "@tanstack/react-table";
import { Link } from "react-router-dom";

const columnHelper = createColumnHelper<V1Beta1User>();

interface getColumnsOptions {
  orgId: string;
  platformUsers: V1Beta1ServiceUser[];
}
export const getColumns: (
  opt: getColumnsOptions
) => ColumnDef<V1Beta1User, any>[] = ({ orgId, platformUsers }) => {
  const platformUsersIdSet = new Set(platformUsers?.map((user) => user?.id));
  return [
    columnHelper.accessor("id", {
      header: "ID",
      //@ts-ignore
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        const serviceUserId = getValue();
        return (
          <Link to={`/organisations/${orgId}/serviceusers/${serviceUserId}`}>
            {serviceUserId}
          </Link>
        );
      },
    }),
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Platform User",
      accessorKey: "",
      cell: ({ row }) =>
        platformUsersIdSet.has(row?.original?.id) ? <CheckCircledIcon /> : null,
    },
    {
      header: "Created At",
      accessorKey: "created_at",
      meta: {
        headerFilter: false,
      },
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),

      footer: (props) => props.column.id,
    },
  ];
};
