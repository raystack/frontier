import { ApsaraColumnDef, Flex } from "@raystack/apsara";
import { V1Beta1Role } from "@raystack/frontier";

import { Link } from "react-router-dom";

export const getColumns: () => ApsaraColumnDef<V1Beta1Role>[] = () => {
  return [
    {
      accessorKey: "id",
      header: "ID",
      filterVariant: "text",
      cell: ({ row, getValue }) => {
        return (
          <Link to={`${encodeURIComponent(row.getValue("id"))}`}>
            {getValue()}
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
      cell: (info) => <Flex>{info.getValue().join(", ")}</Flex>,
      footer: (props) => props.column.id,
    },
  ];
};
