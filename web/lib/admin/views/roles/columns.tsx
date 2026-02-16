import { type DataTableColumnDef, Flex } from "@raystack/apsara";
import type { Role } from "@raystack/proton/frontier";
import styles from "./roles.module.css";
export const getColumns: () => DataTableColumnDef<Role, unknown>[] = () => {
  return [
    {
      accessorKey: "id",
      header: "ID",
      filterVariant: "text",
      cell: ({ getValue }) => getValue(),
    },
    {
      header: "Title",
      accessorKey: "title",
      filterVariant: "text",
      cell: info => info.getValue(),
    },
    {
      header: "Name",
      accessorKey: "name",
      filterVariant: "text",
      cell: info => info.getValue(),
    },
    {
      header: "Permissions",
      accessorKey: "permissions",
      enableColumnFilter: false,
      classNames: {
        cell: styles.permissionsColumn,
      },
      cell: info => (
        <Flex direction="column" gap={1} className={styles.permissionsColumn}>
          {((info.getValue() as string[]) || []).map((p, i) => (
            <span key={i}>{p}</span>
          ))}
        </Flex>
      ),
      footer: props => props.column.id,
    },
  ];
};
