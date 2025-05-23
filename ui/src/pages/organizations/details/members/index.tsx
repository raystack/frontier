import {
  DataTable,
  DataTableQuery,
  DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";
import { useCallback, useContext, useEffect, useState } from "react";
import { api } from "~/api";
import { getColumns } from "./columns";
import type { SearchOrganizationUsersResponseOrganizationUser } from "~/api/frontier";
import UserIcon from "~/assets/icons/users.svg?react";
import { OrganizationContext } from "../contexts/organization-context";
import { AssignRole } from "~/components/assign-role";
import { RemoveMember } from "./remove-member";
import { useRQL } from "~/hooks/useRQL";

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
  const { roles = [], organization, search } = useContext(OrganizationContext);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const organizationId = organization?.id || "";

  const [assignRoleConfig, setAssignRoleConfig] = useState<{
    isOpen: boolean;
    user: SearchOrganizationUsersResponseOrganizationUser | null;
  }>({ isOpen: false, user: null });
  const [removeMemberConfig, setRemoveMemberConfig] = useState<{
    isOpen: boolean;
    user: SearchOrganizationUsersResponseOrganizationUser | null;
  }>({ isOpen: false, user: null });

  const title = `Members | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      const response = await api?.adminServiceSearchOrganizationUsers(
        organizationId,
        { ...apiQuery, search: searchQuery || "" },
      );
      return response?.data;
    },
    [organizationId, searchQuery],
  );

  const { data, setData, loading, query, onTableQueryChange, fetchMore } =
    useRQL<SearchOrganizationUsersResponseOrganizationUser>({
      initialQuery: { offset: 0 },
      key: organizationId,
      dataKey: "org_users",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch tokens:", error),
    });

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  function openAssignRoleDialog(
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) {
    setAssignRoleConfig({ isOpen: true, user });
  }

  function closeAssignRoleDialog() {
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  function openRemoveMemberDialog(
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) {
    setRemoveMemberConfig({ isOpen: true, user });
  }

  function closeRemoveMemberDialog() {
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  const columns = getColumns({
    roles,
    handleAssignRoleAction: openAssignRoleDialog,
    handleRemoveMemberAction: openRemoveMemberDialog,
  });

  async function updateMember(
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) {
    setData((prevMembers) => {
      const updatedMembers = prevMembers.map((member) =>
        member.id === user.id ? user : member,
      );
      return updatedMembers;
    });
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  async function removeMember(
    user: SearchOrganizationUsersResponseOrganizationUser,
  ) {
    setData((prevMembers) => {
      return prevMembers.filter((member) => member.id !== user.id);
    });
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  return (
    <>
      {assignRoleConfig.isOpen && assignRoleConfig.user ? (
        <AssignRole
          roles={roles}
          user={assignRoleConfig.user}
          organizationId={organizationId}
          onRoleUpdate={updateMember}
          onClose={closeAssignRoleDialog}
        />
      ) : null}

      {removeMemberConfig.isOpen && removeMemberConfig.user ? (
        <RemoveMember
          organizationId={organizationId}
          user={removeMemberConfig.user}
          onRemove={removeMember}
          onClose={closeRemoveMemberDialog}
        />
      ) : null}
      <Flex justify="center" className={styles["container"]}>
        <PageTitle title={title} />
        <DataTable
          columns={columns}
          data={data}
          isLoading={loading}
          defaultSort={DEFAULT_SORT}
          mode="server"
          onTableQueryChange={onTableQueryChange}
          onLoadMore={fetchMore}
          query={{ ...query, search: searchQuery }}
        >
          <Flex direction="column" style={{ width: "100%" }}>
            <DataTable.Toolbar />
            <DataTable.Content
              emptyState={<NoMembers />}
              classNames={{
                table: styles["table"],
                root: styles["table-wrapper"],
                header: styles["table-header"],
              }}
            />
          </Flex>
        </DataTable>
      </Flex>
    </>
  );
}
