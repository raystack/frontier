import { type DataTableColumnDef, Flex } from "@raystack/apsara/v1";
import type { V1Beta1Role } from "@raystack/frontier";

import { Link } from "react-router-dom";
import styles from "./roles.module.css";
export const getColumns: () => DataTableColumnDef<
  V1Beta1Role,
  unknown
>[] = () => {
  return [
    {
      accessorKey: "id",
      header: "ID",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`${encodeURIComponent(row.getValue("id"))}`}>
            {getValue() as string}
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
      header: "Name",
      accessorKey: "name",
      filterVariant: "text",
      cell: (info) => info.getValue(),
    },
    {
      header: "Permissions",
      accessorKey: "permissions",
      enableColumnFilter: false,
      classNames: {
        cell: styles.permissionsColumn,
      },
      cell: (info) => <Flex>{(info.getValue() as string[]).join(", ")}</Flex>,
      footer: (props) => props.column.id,
    },
  ];
};
