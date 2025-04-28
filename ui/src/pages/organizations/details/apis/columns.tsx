import { DataTableColumnDef } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { SearchOrganizationServiceUserCredentialsResponseOrganizationServiceUserCredential } from "~/api/frontier";
import { NULL_DATE } from "~/utils/constants";
import styles from "./apis.module.css";

interface ColumnOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export function getColumns(
  options: ColumnOptions,
): DataTableColumnDef<
  SearchOrganizationServiceUserCredentialsResponseOrganizationServiceUserCredential,
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
        return <>{value}</>;
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "serviceuser_title",
      header: "Account",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return <>{value}</>;
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
