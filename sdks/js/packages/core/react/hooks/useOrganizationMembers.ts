import { useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { V1Beta1User } from '~/src';
import { useFrontier } from '../contexts/FrontierContext';

export const useOrganizationMembers = () => {
  const [users, setUsers] = useState([]);
  const [isUsersLoading, setIsUsersLoading] = useState(false);

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
      // @ts-ignore
      setUsers([...users, ...invitedUsers]);
    } catch (err) {
      console.error(err);
    } finally {
      setIsUsersLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    fetchOrganizationUser();
  }, [fetchOrganizationUser, routerState.location.key]);

  const updatedUsers = useMemo(
    () =>
      isUsersLoading
        ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
        : users.length
        ? users
        : [],
    [isUsersLoading, users]
  );
  return {
    isFetching: isUsersLoading,
    members: updatedUsers
  };
};
