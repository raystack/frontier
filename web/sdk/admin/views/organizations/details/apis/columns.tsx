import { Text, type DataTableColumnDef } from "@raystack/apsara";
import dayjs from "dayjs";
import { NULL_DATE } from "../../../../utils/constants";
import styles from "./apis.module.css";
import type {
  SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
  SearchOrganizationServiceUsersResponse_Project
} from "@raystack/proton/frontier";
import { TerminologyEntity } from "../../../../hooks/useAdminTerminology";

interface ColumnOptions {
  groupCountMap: Record<string, Record<string, number>>;
  t: {
    project: TerminologyEntity;
  };
}

export function getColumns(
  options: ColumnOptions,
): DataTableColumnDef<
SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
  unknown
>[] {
  const { t } = options;
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
      header: t.project({ plural: true, case: "capital" }),
      cell: ({ getValue }) => {
        const value =
          getValue() as SearchOrganizationServiceUsersResponse_Project[];
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
