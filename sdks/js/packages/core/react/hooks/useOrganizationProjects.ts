import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import {
  V1Beta1ListProjectsByCurrentUserResponseAccessPair,
  V1Beta1Project
} from '~/src';

interface useOrganizationProjectsProps {
  showInhreitedProjects?: boolean;
  withMemberCount?: boolean;
  allProjects?: boolean;
}

export const useOrganizationProjects = ({
  withMemberCount = false,
  allProjects = false
}: useOrganizationProjectsProps) => {
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);
  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [accessPairs, setAccessPairs] = useState<
    V1Beta1ListProjectsByCurrentUserResponseAccessPair[]
  >([]);

  const { client, activeOrganization: organization } = useFrontier();

  const getProjects = useCallback(
    async (org_id: string) => {
      try {
        setIsProjectsLoading(true);
        const resp = allProjects
          ? await client?.frontierServiceListOrganizationProjects(org_id, {
              with_member_count: withMemberCount
            })
          : await client?.frontierServiceListProjectsByCurrentUser({
              org_id,
              with_permissions: ['update', 'delete'],
              non_inherited: true,
              with_member_count: withMemberCount
            });

        const newProjects = resp?.data?.projects || [];
        // @ts-ignore
        const access_pairs = resp?.data?.access_pairs || [];
        setProjects(newProjects);
        setAccessPairs(access_pairs);
      } catch (err) {
        console.error(err);
      } finally {
        setIsProjectsLoading(false);
      }
    },
    [allProjects, client, withMemberCount]
  );

  const updatedProjects = useMemo(
    () =>
      isProjectsLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : projects.length
        ? projects
        : [],
    [isProjectsLoading, projects]
  );

  const refetch = useCallback(() => {
    if (organization?.id) {
      getProjects(organization?.id);
    }
  }, [getProjects, organization?.id]);

  useEffect(() => {
    refetch();
  }, [refetch]);

  const userAccessOnProject = useMemo(() => {
    return accessPairs.reduce((acc: any, p: any) => {
      const { group_id, permissions } = p;
      acc[group_id] = permissions;
      return acc;
    }, {});
  }, [accessPairs]);

  return {
    isFetching: isProjectsLoading,
    projects: updatedProjects,
    userAccessOnProject,
    refetch: refetch
  };
};
