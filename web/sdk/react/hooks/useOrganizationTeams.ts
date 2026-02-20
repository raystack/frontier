import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, type Group, ListOrganizationGroupsRequestSchema, ListCurrentUserGroupsRequestSchema, type ListCurrentUserGroupsResponse_AccessPair } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

interface useOrganizationTeamsProps {
  withPermissions?: string[];
  showOrgTeams?: boolean;
  withMemberCount?: boolean;
}

export const useOrganizationTeams = ({
  withPermissions = [],
  showOrgTeams = false,
  withMemberCount = false
}: useOrganizationTeamsProps): {
  isFetching: boolean;
  teams: Group[];
  userAccessOnTeam: Record<string, string[]>;
  refetch: () => void;
  error: unknown;
} => {
  const [teams, setTeams] = useState<Group[]>([]);
  const [accessPairs, setAccessPairs] = useState<ListCurrentUserGroupsResponse_AccessPair[]>([]);

  const { activeOrganization: organization } = useFrontier();

  // Organization teams query
  const { data: orgTeamsData, isLoading: isOrgTeamsLoading, error: orgTeamsError, refetch: refetchOrgTeams } = useQuery(
    FrontierServiceQueries.listOrganizationGroups,
    create(ListOrganizationGroupsRequestSchema, { orgId: organization?.id || '', withMemberCount }),
    { enabled: !!organization?.id && showOrgTeams }
  );

  // User teams query  
  const { data: userTeamsData, isLoading: isUserTeamsLoading, error: userTeamsError, refetch: refetchUserTeams } = useQuery(
    FrontierServiceQueries.listCurrentUserGroups,
    create(ListCurrentUserGroupsRequestSchema, { orgId: organization?.id || '', withPermissions, withMemberCount }),
    { enabled: !!organization?.id && !showOrgTeams }
  );

  const isTeamsLoading = showOrgTeams ? isOrgTeamsLoading : isUserTeamsLoading;
  const teamsData = showOrgTeams ? orgTeamsData : userTeamsData;

  useEffect(() => {
    if (teamsData) {
      const { groups = [] } = teamsData;
      setTeams(groups);
      
      // Only set accessPairs for user teams (not org teams)
      if (!showOrgTeams && 'accessPairs' in teamsData) {
        setAccessPairs(teamsData?.accessPairs || []);
      } else {
        setAccessPairs([]);
      }
    }
  }, [teamsData, showOrgTeams]);

  const userAccessOnTeam = useMemo(() => {
    return accessPairs.reduce((acc: Record<string, string[]>, p: ListCurrentUserGroupsResponse_AccessPair) => {
      const { groupId, permissions } = p;
      acc[groupId] = permissions;
      return acc;
    }, {});
  }, [accessPairs]);

  const refetch = useCallback(() => {
    if (showOrgTeams) {
      refetchOrgTeams();
    } else {
      refetchUserTeams();
    }
  }, [showOrgTeams, refetchOrgTeams, refetchUserTeams]);

  return {
    isFetching: isTeamsLoading,
    teams: teams,
    userAccessOnTeam,
    refetch,
    error: showOrgTeams ? orgTeamsError : userTeamsError
  };
};
