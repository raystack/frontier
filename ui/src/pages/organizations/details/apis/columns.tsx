import { Text, type DataTableColumnDef } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import type {
  SearchOrganizationServiceUsersResponseOrganizationServiceUser,
  V1Beta1SearchOrganizationServiceUsersResponseProject,
} from "~/api/frontier";
import { NULL_DATE } from "~/utils/constants";
import styles from "./apis.module.css";

interface ColumnOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export function getColumns(
  options: ColumnOptions,
): DataTableColumnDef<
  SearchOrganizationServiceUsersResponseOrganizationServiceUser,
  unknown
>[] {
  return [
    {
      accessorKey: "title",
      header: "Keys",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return <Text>{value}</Text>;
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "projects",
      header: "Projects",
      cell: ({ getValue }) => {
        const value =
          getValue() as V1Beta1SearchOrganizationServiceUsersResponseProject[];
        const projectNames = value.map((project) => project.title).join(", ");
        return <Text>{projectNames}</Text>;
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "created_at",
      header: "Created on",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableSorting: true,
      enableColumnFilter: true,
      filterType: "date",
      enableHiding: true,
    },
  ];
}
