import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import {
  Outlet,
  useNavigate,
  useOutletContext,
  useParams,
} from "react-router-dom";

import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import PageTitle from "~/components/page-title";
import styles from "./roles.module.css";
import { SheetHeader } from "~/components/sheet/header";
import { FrontierServiceQueries, Role } from "@raystack/proton/frontier";
import { useQuery } from "@connectrpc/connect-query";

type ContextType = { role: Role | null };
export default function RoleList() {
  const { roleId } = useParams();
  const navigate = useNavigate();

  const { data: roles = [], isLoading } = useQuery(
    FrontierServiceQueries.listRoles,
    {},
    {
      select: data => data?.roles ?? [],
    },
  );
  const roleMapByName = reduceByKey(roles ?? [], "id");

  function onClose() {
    navigate("/roles");
  }
  function onRowClick(role: Role) {
    navigate(`${encodeURIComponent(role.id ?? "")}`);
  }

  const columns = getColumns();
  return (
    <DataTable
      onRowClick={onRowClick}
      data={roles}
      columns={columns}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
      isLoading={isLoading}>
      <Flex direction="column">
        <PageTitle title="Roles" />
        <RolesHeader />
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{
            root: styles.tableRoot,
            table: styles.table,
          }}
        />
        <Sheet open={roleId !== undefined}>
          <Sheet.Content className={styles.sheetContent}>
            <SheetHeader
              title="Role Details"
              onClick={onClose}
              data-test-id="role-details-header"
            />
            <Flex className={styles.sheetContentBody}>
              <Outlet
                context={{
                  role: roleId ? roleMapByName[roleId] : null,
                }}
              />
            </Flex>
          </Sheet.Content>
        </Sheet>
      </Flex>
    </DataTable>
  );
}

export function useRole() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 role created" />
);
