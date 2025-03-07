import { Avatar, DataTableColumnDef, Flex, Text } from "@raystack/apsara/v1";
import { V1Beta1Organization } from "@raystack/frontier";
import dayjs from "dayjs";
import styles from "./list.module.css";

export const getColumns = (): DataTableColumnDef<
  V1Beta1Organization,
  unknown
>[] => {
  return [
    {
      accessorKey: "title",
      header: "Name",
      columnType: "text",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: ({ row }) => {
        return (
          <Flex gap={4} align="center">
            <Avatar /> <Text>{row.original.title}</Text>
          </Flex>
        );
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "created_by",
      header: "Creator",
      columnType: "text",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableSorting: true,
    },
    {
      accessorKey: "plan_name",
      header: "Plan",
      columnType: "text",
      cell: ({ getValue }) => {
        // TODO: update as select
        return getValue();
      },
      enableColumnFilter: true,
      enableHiding: true,
    },
    {
      accessorKey: "subscription_cycle_end_at",
      header: "Cycle ends on",
      columnType: "date",
      cell: ({ getValue }) => {
        // TODO: hanlde data zero value
        return dayjs(getValue() as string).format("YYYY-MM-DD");
      },
      enableHiding: true,
    },
    {
      accessorKey: "country",
      header: "Country",
      columnType: "text",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableHiding: true,
    },
    {
      accessorKey: "payment_mode",
      header: "Payment mode",
      columnType: "text",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableHiding: true,
      defaultHidden: true,
    },
    {
      accessorKey: "subscription_state",
      header: "Status",
      columnType: "text",
      cell: ({ getValue }) => {
        // TODO: update as select
        return getValue();
      },
      enableHiding: true,
      defaultHidden: true,
    },
    {
      accessorKey: "created_at",
      header: "Created On",
      columnType: "date",
      cell: ({ getValue }) => {
        return dayjs(getValue() as string).format("YYYY-MM-DD");
      },
      enableHiding: true,
      defaultHidden: true,
      enableSorting: true,
    },
  ];
};
