import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";
import { useCallback, useEffect, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import { SearchOrganizationUsersResponseOrganizationUser } from "~/api/frontier";
import UserIcon from "~/assets/icons/users.svg?react";

const DEFAULT_SORT: DataTableSort = { name: "org_joined_at", order: "desc" };

const NoMembers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Member found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

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
      if (!organizationId) return;
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
          <DataTable.Content emptyState={<NoMembers />} />
        </Flex>
      </DataTable>
    </Flex>
  );
}
