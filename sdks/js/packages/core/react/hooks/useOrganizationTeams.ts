import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { V1Beta1Group } from '~/src';

interface useOrganizationTeamsProps {
  withPermissions?: string[];
  showOrgTeams?: boolean;
}

export const useOrganizationTeams = ({
  withPermissions = [],
  showOrgTeams = false
}: useOrganizationTeamsProps) => {
  const [teams, setTeams] = useState<V1Beta1Group[]>([]);
  const [isTeamsLoading, setIsTeamsLoading] = useState(false);
  const [accessPairs, setAccessPairs] = useState([]);

  const { client, activeOrganization: organization } = useFrontier();

  const getTeams = useCallback(async () => {
    if (!organization?.id) return;
    try {
      setIsTeamsLoading(true);
      const {
        // @ts-ignore
        data: { groups = [], access_pairs = [] }
      } = showOrgTeams
        ? await client?.frontierServiceListOrganizationGroups(organization?.id)
        : await client?.frontierServiceListCurrentUserGroups({
            // @ts-ignore
            org_id: organization?.id,
            withPermissions
          });
      setTeams(groups);
      setAccessPairs(access_pairs);
    } catch (err) {
      console.error(err);
    } finally {
      setIsTeamsLoading(false);
    }
  }, [client, organization?.id, showOrgTeams]);

  useEffect(() => {
    getTeams();
  }, [client, getTeams, organization?.id]);

  const updatedTeams = useMemo(
    () =>
      isTeamsLoading
        ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
        : teams.length
        ? teams
        : [],
    [isTeamsLoading, teams]
  );

  const userAccessOnTeam = useMemo(() => {
    return accessPairs.reduce((acc: any, p: any) => {
      const { group_id, permissions } = p;
      acc[group_id] = permissions;
      return acc;
    }, {});
  }, [accessPairs]);

  return {
    isFetching: isTeamsLoading,
    teams: updatedTeams,
    userAccessOnTeam,
    refetch: getTeams
  };
};
