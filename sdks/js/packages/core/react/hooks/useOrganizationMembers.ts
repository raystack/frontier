import { useCallback, useEffect, useState } from 'react';
import { V1Beta1User, V1Beta1Role, V1Beta1Invitation } from '~/src';
import { PERMISSIONS } from '~/utils';
import { useFrontier } from '../contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, ListOrganizationUsersRequestSchema, ListRolesRequestSchema, ListOrganizationInvitationsRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';


export type MemberWithInvite = V1Beta1User & V1Beta1Invitation & {invited?: boolean}

export interface UseOrganizationMembersReturn {
  isFetching: boolean;
  members: MemberWithInvite[];
  memberRoles: Record<string, V1Beta1Role[]>;
  roles: V1Beta1Role[];
  refetch: () => void;
  error: unknown;
}

export const useOrganizationMembers = ({ showInvitations = false }): UseOrganizationMembersReturn => {
  const [users, setUsers] = useState<V1Beta1User[]>([]);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [invitations, setInvitations] = useState<MemberWithInvite[]>([]);

  const [memberRoles, setMemberRoles] = useState<Record<string, V1Beta1Role[]>>({});

  const { activeOrganization: organization } = useFrontier();

  const { data: organizationUsersData, isLoading: isUsersLoading, error: usersError, refetch: refetchUsers } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    create(ListOrganizationUsersRequestSchema, {
      id: organization?.id || '',
      withRoles: true
    }),
    { enabled: !!organization?.id }
  );

  useEffect(() => {
    if (organizationUsersData) {
      const { users, rolePairs } = organizationUsersData;
      setUsers(users || []);
      setMemberRoles(
        (rolePairs || []).reduce((previous: Record<string, V1Beta1Role[]>, mr: { userId: string; roles: V1Beta1Role[] }) => {
          return { ...previous, [mr.userId]: mr.roles };
        }, {})
      );
    }
  }, [organizationUsersData]);

  const { data: rolesData, isLoading: isRolesLoading, error: rolesError, refetch: refetchRoles } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.OrganizationNamespace]
    }),
    { enabled: !!organization?.id }
  );

  useEffect(() => {
    if (rolesData) {
      setRoles(rolesData.roles || []);
    }
  }, [rolesData]);

  const { data: invitationsData, isLoading: isInvitationsLoading, error: invitationsError, refetch: refetchInvitations } = useQuery(
    FrontierServiceQueries.listOrganizationInvitations,
    create(ListOrganizationInvitationsRequestSchema, {
      orgId: organization?.id || ''
    }),
    { enabled: !!organization?.id && showInvitations }
  );

  useEffect(() => {
    if (invitationsData) {
      const invitedUsers: MemberWithInvite[] = (invitationsData.invitations || []).map((user: V1Beta1User) => ({
        ...user,
        invited: true
      }));
      setInvitations(invitedUsers);
    }
  }, [invitationsData]);


  const isFetching = isUsersLoading || isInvitationsLoading || isRolesLoading;
  const hasError = usersError || rolesError || invitationsError;

  const updatedUsers = [...users, ...invitations];

  const refetch = useCallback(() => {
    // Trigger refetch of all queries
    refetchUsers();
    refetchRoles();
    refetchInvitations();
  }, [refetchUsers, refetchRoles, refetchInvitations]);

  return {
    isFetching,
    members: updatedUsers,
    memberRoles,
    roles,
    refetch,
    error: hasError
  };
};
