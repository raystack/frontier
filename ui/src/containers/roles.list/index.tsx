import { EmptyState, Flex, DataTable, Sheet } from "@raystack/apsara";
import { useEffect, useState } from "react";
import {
  Outlet,
  useNavigate,
  useOutletContext,
  useParams,
} from "react-router-dom";

import type { V1Beta1Role } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";
import PageTitle from "~/components/page-title";
import styles from "./roles.module.css";
import { SheetHeader } from "~/components/sheet/header";

type ContextType = { role: V1Beta1Role | null };
export default function RoleList() {
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [isRolesLoading, setIsRolesLoading] = useState(false);

  useEffect(() => {
    async function getRoles() {
      setIsRolesLoading(true);
      try {
        const res = await api?.frontierServiceListRoles();
        const roles = res?.data?.roles ?? [];
        setRoles(roles);
      } catch (err) {
        console.log(err);
      } finally {
        setIsRolesLoading(false);
      }
    }
    getRoles();
  }, []);
  let { roleId } = useParams();
  const roleMapByName = reduceByKey(roles ?? [], "id");
  const navigate = useNavigate();

  function onClose() {
    navigate("/roles");
  }

  const columns = getColumns();
  return (
    <DataTable
      data={roles}
      columns={columns}
      mode="client"
      defaultSort={{ name: "title", order: "asc" }}
      isLoading={isRolesLoading}
    >
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
            <SheetHeader title="Role Details" onClick={onClose} />
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
