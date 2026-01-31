import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import { useCallback, useState } from "react";

import { reduceByKey } from "../../utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { PageTitle } from "../../components/PageTitle";
import styles from "./roles.module.css";
import { SheetHeader } from "../../components/SheetHeader";
import { FrontierServiceQueries, Role } from "@raystack/proton/frontier";
import { useQuery } from "@connectrpc/connect-query";

import RoleDetails from "./details";

export type RolesPageProps = {
  selectedRoleId?: string;
  onSelectRole?: (roleId: string) => void;
  onCloseDetail?: () => void;
  appName?: string;
};

export default function RolesPage({
  selectedRoleId: controlledRoleId,
  onSelectRole,
  onCloseDetail,
  appName,
}: RolesPageProps = {}) {
  const [internalRoleId, setInternalRoleId] = useState<string | undefined>();

  const selectedRoleId = controlledRoleId ?? internalRoleId;
  const handleClose = useCallback(
    () => (onCloseDetail ? onCloseDetail() : setInternalRoleId(undefined)),
    [onCloseDetail]
  );
  const handleRowClick = useCallback(
    (role: Role) =>
      onSelectRole ? onSelectRole(role.id ?? "") : setInternalRoleId(role.id ?? undefined),
    [onSelectRole]
  );

  const {
    data: roles = [],
    isLoading,
    error,
    isError,
  } = useQuery(
    FrontierServiceQueries.listRoles,
    {},
    {
      select: data => data?.roles ?? [],
    },
  );
  const roleMapByName = reduceByKey(roles ?? [], "id");

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Roles"
        subHeading={
          error?.message ||
          "Something went wrong while loading roles. Please try again."
        }
      />
    );
  }

  const columns = getColumns();
  return (
    <DataTable
      onRowClick={handleRowClick}
      data={roles}
      columns={columns}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
      isLoading={isLoading}>
      <Flex direction="column">
        <PageTitle title="Roles" appName={appName} />
        <RolesHeader />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{
            root: styles.tableRoot,
            table: styles.table,
          }}
        />
        <Sheet open={selectedRoleId !== undefined}>
          <Sheet.Content className={styles.sheetContent}>
            <SheetHeader
              title="Role Details"
              onClick={handleClose}
              data-testid="role-details-header"
            />
            <Flex className={styles.sheetContentBody}>
              <RoleDetails
                role={selectedRoleId ? roleMapByName[selectedRoleId] ?? null : null}
              />
            </Flex>
          </Sheet.Content>
        </Sheet>
      </Flex>
    </DataTable>
  );
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 role created" />
);
