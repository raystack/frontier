import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";
import { useCallback, useEffect, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import { SearchOrganizationUsersResponseOrganizationUser } from "~/api/frontier";

const DEFAULT_SORT: DataTableSort = { name: "org_joined_at", order: "desc" };

export function OrganizationMembersPage() {
  const { organization } = useOutletContext<OutletContext>();
  const organizationId = organization.id || "";

  const [data, setData] = useState<
    SearchOrganizationUsersResponseOrganizationUser[]
  >([]);
  const [isDataLoading, setIsDataLoading] = useState(false);

  const title = `Members | ${organization.title} | Organizations`;

  const fetchMembers = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      try {
        setIsDataLoading(true);
        const response = await api?.adminServiceSearchOrganizationUsers(
          organizationId,
          apiQuery,
        );
        const members = response.data.org_users || [];
        setData((prev) => [...prev, ...members]);
      } catch (error) {
        console.error(error);
      } finally {
        setIsDataLoading(false);
      }
    },
    [organizationId],
  );

  useEffect(() => {
    fetchMembers({ offset: 0, sort: [DEFAULT_SORT] });
  }, [fetchMembers]);

  const columns = getColumns();

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isDataLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content />
        </Flex>
      </DataTable>
    </Flex>
  );
}
