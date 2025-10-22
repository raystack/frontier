import { useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListOrganizationGroupsRequestSchema,
  ListCurrentUserGroupsRequestSchema,
  Group
} from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';

interface useOrganizationTeamsProps {
  withPermissions?: string[];
  showOrgTeams?: boolean;
  withMemberCount?: boolean;
}

export const useOrganizationTeams = ({
  withPermissions = [],
  showOrgTeams = false,
  withMemberCount = false
}: useOrganizationTeamsProps) => {
  const { activeOrganization: organization } = useFrontier();

  const {
    data: orgTeamsData,
    isLoading: isOrgTeamsLoading,
    refetch: refetchOrgTeams
  } = useQuery(
    FrontierServiceQueries.listOrganizationGroups,
    create(ListOrganizationGroupsRequestSchema, {
      orgId: organization?.id ?? '',
      withMemberCount
    }),
    {
      enabled: showOrgTeams && !!organization?.id
    }
  );

  const {
    data: userTeamsData,
    isLoading: isUserTeamsLoading,
    refetch: refetchUserTeams
  } = useQuery(
    FrontierServiceQueries.listCurrentUserGroups,
    create(ListCurrentUserGroupsRequestSchema, {
      orgId: organization?.id ?? '',
      withPermissions,
      withMemberCount
    }),
    {
      enabled: !showOrgTeams && !!organization?.id
    }
  );

  const teams = useMemo(() => {
    if (showOrgTeams) {
      return (orgTeamsData?.groups ?? []) as Group[];
    }
    return (userTeamsData?.groups ?? []) as Group[];
  }, [showOrgTeams, orgTeamsData?.groups, userTeamsData?.groups]);

  const userAccessOnTeam = useMemo(() => {
    const accessPairs = userTeamsData?.accessPairs ?? [];
    return accessPairs.reduce((acc: Record<string, string[]>, p) => {
      const groupId = p.groupId ?? '';
      const permissions = p.permissions ?? [];
      acc[groupId] = permissions;
      return acc;
    }, {});
  }, [userTeamsData?.accessPairs]);

  const isFetching = showOrgTeams ? isOrgTeamsLoading : isUserTeamsLoading;

  const refetch = () => {
    if (showOrgTeams) {
      refetchOrgTeams();
    } else {
      refetchUserTeams();
    }
  };

  return {
    isFetching,
    teams,
    userAccessOnTeam,
    refetch
  };
};
