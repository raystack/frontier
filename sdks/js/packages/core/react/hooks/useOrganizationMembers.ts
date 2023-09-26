import { useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { V1Beta1User } from '~/src';
import { useFrontier } from '../contexts/FrontierContext';

export const useOrganizationMembers = ({ showInvitations = false }) => {
  const [users, setUsers] = useState([]);
  const [invitations, setInvitations] = useState([]);

  const [isUsersLoading, setIsUsersLoading] = useState(false);
  const [isInvitationsLoading, setIsInvitationsLoading] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();

  const fetchOrganizationUser = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsUsersLoading(true);
      const {
        // @ts-ignore
        data: { users }
      } = await client?.frontierServiceListOrganizationUsers(organization?.id);
      setUsers(users);
    } catch (err) {
      console.error(err);
    } finally {
      setIsUsersLoading(false);
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
  }, [fetchOrganizationUser, routerState.location.key]);

  useEffect(() => {
    if (showInvitations) {
      fetchInvitations();
    }
  }, [showInvitations, fetchInvitations, routerState.location.key]);

  const isFetching = isUsersLoading || isInvitationsLoading;

  const updatedUsers = useMemo(() => {
    const totalUsers = [...users, ...invitations];
    return isFetching
      ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
      : totalUsers.length
      ? totalUsers
      : [];
  }, [invitations, isFetching, users]);

  return {
    isFetching,
    members: updatedUsers
  };
};
