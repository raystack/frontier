import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import {
  V1Beta1ListOrganizationUsersResponseRolePair,
  V1Beta1Organization,
  V1Beta1Role,
  V1Beta1User,
} from "@raystack/frontier";
import { OrganizationsUsersHeader } from "./header";
import { getColumns } from "./columns";
import { reduceByKey } from "~/utils/helper";
import { PERMISSIONS } from "~/utils/constants";

type ContextType = { user: V1Beta1User | null };

export default function OrganisationUsers() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [users, setOrgUsers] = useState<V1Beta1User[]>([]);
  const [rolePairs, setRolePairs] = useState<
    V1Beta1ListOrganizationUsersResponseRolePair[]
  >([]);
  const [isRolesLoading, setIsRolesLoading] = useState(false);
  const [isUsersLoading, setIsUsersLoading] = useState(false);
  const [isOrgLoading, setIsOrgLoading] = useState(false);

  const [roles, setRoles] = useState<V1Beta1Role[]>([]);

  const getRoles = useCallback(
    async (ordId: string) => {
      try {
        setIsRolesLoading(true);
        const [orgRolesResp, allRolesResp] = await Promise.all([
          client?.frontierServiceListOrganizationRoles(ordId, {
            scopes: [PERMISSIONS.OrganizationNamespace],
          }),
          client?.frontierServiceListRoles({
            scopes: [PERMISSIONS.OrganizationNamespace],
          }),
        ]);
        setRoles([
          ...(orgRolesResp?.data?.roles || []),
          ...(allRolesResp?.data?.roles || []),
        ]);
      } catch (err) {
        console.error(err);
      } finally {
        setIsRolesLoading(false);
      }
    },
    [client]
  );

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisationId}`,
        name: `${organisation?.name}`,
      },
      {
        href: ``,
        name: `Organizations Users`,
      },
    ],
  };

  const getOrganizationUser = useCallback(
    async (orgId: string) => {
      try {
        setIsUsersLoading(true);
        const [usersResp, invitationResp] = await Promise.all([
          client?.frontierServiceListOrganizationUsers(orgId, {
            with_roles: true,
          }),
          client?.frontierServiceListOrganizationInvitations(orgId),
        ]);
        const userList = usersResp?.data?.users || [];
        const role_pairs = usersResp?.data?.role_pairs || [];
        const invitedUsers =
          invitationResp?.data?.invitations?.map((user) => ({
            isInvited: true,
            email: user?.user_id,
            ...user,
          })) || [];
        setOrgUsers([...userList, ...invitedUsers]);
        setRolePairs(role_pairs);
      } catch (err) {
        console.error(err);
      } finally {
        setIsUsersLoading(false);
      }
    },
    [client]
  );

  useEffect(() => {
    async function getOrganization(orgId: string) {
      try {
        setIsOrgLoading(true);
        const {
          // @ts-ignore
          data: { organization },
        } = await client?.frontierServiceGetOrganization(orgId) ?? {};
        setOrganisation(organization);
      } catch (err) {
        console.error(err);
      } finally {
        setIsOrgLoading(false);
      }
    }
    if (organisationId) {
      getOrganization(organisationId);
      getOrganizationUser(organisationId);
      getRoles(organisationId);
    }
  }, [client, getOrganizationUser, getRoles, organisationId]);

  const tableStyle = users?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const rolesMapByUserId = reduceByKey(rolePairs, "user_id");

  const columns = getColumns({
    users,
    orgId: organisationId || "",
    userRolesMap: rolesMapByUserId,
    roles,
    refetchUsers: () =>
      organisationId ? getOrganizationUser(organisationId) : {},
  });

  const isLoading = isRolesLoading || isUsersLoading || isOrgLoading;

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={users}
        // @ts-ignore
        columns={columns}
        isLoading={isLoading}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsUsersHeader
            header={pageHeader}
            orgId={organisationId}
          />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useUser() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>No users created</h3>
    <div className="pera">Try creating a new user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
