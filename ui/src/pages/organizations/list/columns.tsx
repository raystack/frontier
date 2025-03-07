import { Avatar, DataTableColumnDef, Flex, Text } from "@raystack/apsara/v1";
import { V1Beta1Organization } from "@raystack/frontier";
import dayjs from "dayjs";

export const getColumns = (): DataTableColumnDef<
  V1Beta1Organization,
  unknown
>[] => {
  return [
    {
      accessorKey: "name",
      header: "Name",
      columnType: "text",
      cell: ({ row }) => {
        return (
          <Flex>
            <Avatar /> <Text>{row.original.title}</Text>
          </Flex>
        );
      },
    },
    {
      accessorKey: "created_by",
      header: "Creator",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
    },
    {
      accessorKey: "plan_name",
      header: "Plan",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
      enableHiding: true,
    },
    {
      accessorKey: "plan_name",
      header: "Cycle ends on",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
      enableHiding: true,
    },
    {
      accessorKey: "plan_name",
      header: "Country",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
      enableHiding: true,
    },
    {
      accessorKey: "plan_name",
      header: "Payment mode",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
      enableHiding: true,
      defaultHidden: true,
    },
    {
      accessorKey: "Status",
      header: "Payment mode",
      columnType: "text",
      cell: ({ row }) => {
        return row.original.title;
      },
      enableHiding: true,
      defaultHidden: true,
    },
    {
      accessorKey: "created_at",
      header: "Created On",
      columnType: "date",
      cell: ({ row }) => {
        return dayjs(row.original.created_at).format("YYYY-MM-DD");
      },
      enableHiding: true,
      defaultHidden: true,
    },
  ];
};
