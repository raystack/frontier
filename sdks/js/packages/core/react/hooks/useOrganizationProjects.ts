import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';

export const useOrganizationProjects = () => {
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);
  const [projects, setProjects] = useState([]);
  const [accessPairs, setAccessPairs] = useState([]);

  const { client, activeOrganization: organization } = useFrontier();

  const getProjects = useCallback(async () => {
    try {
      setIsProjectsLoading(true);
      const {
        // @ts-ignore
        data: { projects = [], access_pairs = [] }
      } = await client?.frontierServiceListProjectsByCurrentUser({
        // @ts-ignore
        org_id: organization?.id,
        withPermissions: ['update', 'delete']
      });
      setProjects(projects);
      setAccessPairs(access_pairs);
    } catch (err) {
      console.error(err);
    } finally {
      setIsProjectsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getProjects();
  }, [client, getProjects, organization?.id]);

  const updatedProjects = useMemo(
    () =>
      isProjectsLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : projects.length
        ? projects
        : [],
    [isProjectsLoading, projects]
  );

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
    userAccessOnProject
  };
};
