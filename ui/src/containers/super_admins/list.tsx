import { DataTable, EmptyState } from "@raystack/apsara";
import { getColumns } from "./columns";
import styles from "./super_admins.module.css";
import { useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export function SuperAdminList() {
  const {
    data: platformUsersData,
    isLoading,
    error,
    isError,
  } = useQuery(AdminServiceQueries.listPlatformUsers, {}, {
    staleTime: Infinity,
  });

  const columns = getColumns();
  const data = [
    ...(platformUsersData?.users || []),
    ...(platformUsersData?.serviceusers || []),
  ];

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Admins"
        subHeading={
          error?.message ||
          "Something went wrong while loading Admins. Please try again."
        }
      />
    );
  }

  return (
    <DataTable
      data={data || []}
      defaultSort={{ name: "email", order: "asc" }}
      columns={columns}
      mode="client"
      isLoading={isLoading}
    >
      <DataTable.Content classNames={{ root: styles.tableRoot }} />
    </DataTable>
  );
}
