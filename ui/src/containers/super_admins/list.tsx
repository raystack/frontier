import { DataTable } from "@raystack/apsara/v1";
import { useContext } from "react";
import { AppContext } from "~/contexts/App";
import { getColumns } from "./columns";
import styles from "./super_admins.module.css";

export function SuperAdminList() {
  const { platformUsers } = useContext(AppContext);

  const columns = getColumns();
  const data = [
    ...(platformUsers?.users || []),
    ...(platformUsers?.serviceusers || []),
  ];
  return (
    <DataTable
      data={data || []}
      defaultSort={{ name: "email", order: "asc" }}
      columns={columns}
      mode="client"
    >
      <DataTable.Content classNames={{ root: styles.tableRoot }} />
    </DataTable>
  );
}
