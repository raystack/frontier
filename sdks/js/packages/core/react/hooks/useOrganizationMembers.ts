import { useCallback, useEffect, useMemo, useState } from 'react';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { PERMISSIONS } from '~/utils';
import { useFrontier } from '../contexts/FrontierContext';

export const useOrganizationMembers = ({ showInvitations = false }) => {
  const [users, setUsers] = useState([]);
  const [roles, setRoles] = useState([]);
  const [invitations, setInvitations] = useState([]);

  const [isUsersLoading, setIsUsersLoading] = useState(false);
  const [isRolesLoading, setIsRolesLoading] = useState(false);
  const [isInvitationsLoading, setIsInvitationsLoading] = useState(false);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});

  const { client, activeOrganization: organization } = useFrontier();

  const fetchOrganizationUser = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsUsersLoading(true);
      const {
        // @ts-ignore
        data: { users, role_pairs }
      } = await client?.frontierServiceListOrganizationUsers(organization?.id, {
        withRoles: true
      });
      setUsers(users);
      setMemberRoles(
        role_pairs.reduce((previous: any, mr: any) => {
          return { ...previous, [mr.user_id]: mr.roles };
        }, {})
      );
    } catch (err) {
      console.error(err);
    } finally {
      setIsUsersLoading(false);
    }
  }, [client, organization?.id]);

  const fetchOrganizationRoles = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsRolesLoading(true);
      const {
        // @ts-ignore
        data: { roles }
      } = await client?.frontierServiceListRoles({
        state: 'enabled',
        scopes: [PERMISSIONS.OrganizationNamespace]
      });
      setRoles(roles);
    } catch (err) {
      console.error(err);
    } finally {
      setIsRolesLoading(false);
    }
  }, [client, organization?.id]);

  const fetchInvitations = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsInvitationsLoading(true);

      const {
        // @ts-ignore
        data: { invitations }
      } = await client?.frontierServiceListOrganizationInvitations(
        organization?.id
      );
      const invitedUsers = invitations.map((user: V1Beta1User) => ({
        ...user,
        invited: true
      }));
      setInvitations(invitedUsers);
    } catch (err) {
      console.error(err);
    } finally {
      setIsInvitationsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [fetchOrganizationUser]);

  useEffect(() => {
    fetchOrganizationRoles();
  }, [fetchOrganizationRoles]);

  useEffect(() => {
    if (showInvitations) {
      fetchInvitations();
    }
  }, [showInvitations, fetchInvitations]);

  const isFetching = isUsersLoading || isInvitationsLoading;

  const updatedUsers = useMemo(() => {
    const totalUsers = [...users, ...invitations];
    return isFetching
      ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
      : totalUsers.length
      ? totalUsers
      : [];
  }, [invitations, isFetching, users]);

  const refetch = useCallback(() => {
    fetchOrganizationUser();
    if (showInvitations) {
      fetchInvitations();
    }
  }, [fetchInvitations, fetchOrganizationUser, showInvitations]);

  return {
    isFetching,
    members: updatedUsers,
    memberRoles,
    roles,
    refetch
  };
};
