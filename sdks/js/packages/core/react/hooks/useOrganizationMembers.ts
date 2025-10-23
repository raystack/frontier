import { useCallback, useEffect, useState } from 'react';
import { V1Beta1User, V1Beta1Role, V1Beta1Invitation } from '~/src';
import { PERMISSIONS } from '~/utils';
import { useFrontier } from '../contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';


export type MemberWithInvite = V1Beta1User & V1Beta1Invitation & {invited?: boolean}


export const useOrganizationMembers = ({ showInvitations = false }): {
  isFetching: boolean;
  members: MemberWithInvite[];
  memberRoles: Record<string, V1Beta1Role[]>;
  roles: V1Beta1Role[];
  refetch: () => void;
  error: unknown;
} => {
  const [users, setUsers] = useState<V1Beta1User[]>([]);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [invitations, setInvitations] = useState<MemberWithInvite[]>([]);

  const [memberRoles, setMemberRoles] = useState<Record<string, V1Beta1Role[]>>({});

  const { activeOrganization: organization } = useFrontier();

  const { data: organizationUsersData, isLoading: isUsersLoading, error: usersError } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    { id: organization?.id || '', withRoles: true },
    { enabled: !!organization?.id }
  );

  useEffect(() => {
    if (organizationUsersData) {
      const { users, rolePairs } = organizationUsersData;
      setUsers(users || []);
      setMemberRoles(
        (rolePairs || []).reduce((previous: any, mr: any) => {
          return { ...previous, [mr.userId]: mr.roles };
        }, {})
      );
    }
  }, [organizationUsersData]);

  const { data: rolesData, isLoading: isRolesLoading, error: rolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    { state: 'enabled', scopes: [PERMISSIONS.OrganizationNamespace] },
    { enabled: !!organization?.id }
  );

  useEffect(() => {
    if (rolesData) {
      setRoles(rolesData.roles || []);
    }
  }, [rolesData]);

  const { data: invitationsData, isLoading: isInvitationsLoading, error: invitationsError } = useQuery(
    FrontierServiceQueries.listOrganizationInvitations,
    { orgId: organization?.id || '' },
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
    // All data is now automatically refetched via useQuery
    // This function is kept for backward compatibility
  }, []);

  return {
    isFetching,
    members: updatedUsers,
    memberRoles,
    roles,
    refetch,
    error: hasError
  };
};
