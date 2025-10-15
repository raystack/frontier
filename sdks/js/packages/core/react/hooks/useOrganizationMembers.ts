import { useCallback, useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { V1Beta1User, V1Beta1Invitation } from '~/src';
import { PERMISSIONS } from '~/utils';
import { useFrontier } from '../contexts/FrontierContext';


export type MemberWithInvite = V1Beta1User & V1Beta1Invitation & {invited?: boolean}



export const useOrganizationMembers = ({ showInvitations = false }) => {
  const { activeOrganization: organization } = useFrontier();

  // List organization users
  const { data: usersData, isLoading: isUsersLoading, refetch: refetchUsers } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    { orgId: organization?.id ?? '', withRoles: true },
    { enabled: !!organization?.id }
  );

  // List roles
  const { data: rolesData, isLoading: isRolesLoading } = useQuery(
    FrontierServiceQueries.listRoles,
    { state: 'enabled', scopes: [PERMISSIONS.OrganizationNamespace] },
    { enabled: !!organization?.id }
  );

  // List organization invitations
  const { data: invitationsData, isLoading: isInvitationsLoading, refetch: refetchInvitations } = useQuery(
    FrontierServiceQueries.listOrganizationInvitations,
    { orgId: organization?.id ?? '' },
    { enabled: !!organization?.id && showInvitations }
  );

  const users = usersData?.users || [];
  const roles = rolesData?.roles || [];
  const invitations = useMemo(() => {
    const invites = invitationsData?.invitations || [];
    return invites.map((user: V1Beta1User) => ({
      ...user,
      invited: true
    })) as MemberWithInvite[];
  }, [invitationsData?.invitations]);

  const memberRoles = useMemo(() => {
    const rolePairs = usersData?.rolePairs || [];
    return rolePairs.reduce((previous: any, mr: any) => {
      return { ...previous, [mr.userId]: mr.roles };
    }, {});
  }, [usersData?.rolePairs]);

  const isFetching = isUsersLoading || isInvitationsLoading || isRolesLoading;
  const updatedUsers = [...users, ...invitations];

  const refetch = useCallback(() => {
    refetchUsers();
    if (showInvitations) {
      refetchInvitations();
    }
  }, [refetchUsers, refetchInvitations, showInvitations]);

  return {
    isFetching,
    members: updatedUsers,
    memberRoles,
    roles,
    refetch
  };
};
