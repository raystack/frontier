import { CheckCircledIcon } from "@radix-ui/react-icons";
import { ApsaraColumnDef } from "@raystack/apsara";
import { V1Beta1ServiceUser, V1Beta1User } from "@raystack/frontier";
import { Link } from "react-router-dom";

interface getColumnsOptions {
  orgId: string;
  platformUsers: V1Beta1ServiceUser[];
}
export const getColumns: (
  opt: getColumnsOptions
) => ApsaraColumnDef<V1Beta1User>[] = ({ orgId, platformUsers }) => {
  const platformUsersIdSet = new Set(platformUsers?.map((user) => user?.id));
  return [
    {
      id: "id",
      header: "ID",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        const serviceUserId = getValue();
        return (
          <Link to={`/organisations/${orgId}/serviceusers/${serviceUserId}`}>
            {serviceUserId}
          </Link>
        );
      },
    },
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
