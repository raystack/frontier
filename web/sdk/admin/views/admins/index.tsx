import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { getColumns } from "./columns";
import styles from "./admins.module.css";
import { useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { PageHeader } from "../../components/PageHeader";
import { useTerminology } from "../../hooks/useTerminology";

const pageHeader = {
  title: "Super Admins",
  breadcrumb: [],
};

const NoAdmins = () => {
  return (
    <EmptyState
      icon={<ExclamationTriangleIcon />}
      heading="No Admins Found"
      subHeading="No platform users or service users found."
    />
  );
};

export type AdminsViewProps = {
  onNavigateToOrg?: (orgId: string) => void;
};

export default function AdminsView({ onNavigateToOrg }: AdminsViewProps = {}) {
  const t = useTerminology();
  const {
    data: platformUsersData,
    isLoading,
    error,
    isError,
  } = useQuery(AdminServiceQueries.listPlatformUsers, {}, {
    staleTime: Infinity,
  });

  const columns = getColumns({ onNavigateToOrg, t });
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
      <Flex direction="column" className={styles.tableWrapper}>
        <PageHeader
          title={pageHeader.title}
          breadcrumb={pageHeader.breadcrumb}
          className={styles.header}
        />
        <DataTable.Content
          classNames={{ root: styles.tableRoot }}
          emptyState={<NoAdmins />}
        />
      </Flex>
    </DataTable>
  );
}
