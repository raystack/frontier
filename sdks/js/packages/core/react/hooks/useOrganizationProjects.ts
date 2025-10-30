import { useCallback, useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, ListOrganizationProjectsRequestSchema, ListProjectsByCurrentUserRequestSchema, type Project, type ListProjectsByCurrentUserResponse_AccessPair } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../contexts/FrontierContext';

interface useOrganizationProjectsProps {
  showInhreitedProjects?: boolean;
  withMemberCount?: boolean;
  allProjects?: boolean;
}

export interface UseOrganizationProjectsReturn {
  isFetching: boolean;
  projects: Project[];
  userAccessOnProject: Record<string, string[]>;
  refetch: () => void;
  error: unknown;
}

export const useOrganizationProjects = ({
  withMemberCount = false,
  allProjects = false
}: useOrganizationProjectsProps): UseOrganizationProjectsReturn => {
  const { activeOrganization: organization } = useFrontier();

  // Query for organization projects (all projects)
  const { 
    data: orgProjectsData, 
    isLoading: isOrgProjectsLoading, 
    error: orgProjectsError,
    refetch: refetchOrgProjects 
  } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, { 
      orgId: organization?.id || '',
      withMemberCount 
    }),
    { enabled: !!organization?.id && allProjects }
  );

  // Query for current user projects
  const { 
    data: userProjectsData, 
    isLoading: isUserProjectsLoading, 
    error: userProjectsError,
    refetch: refetchUserProjects 
  } = useQuery(
    FrontierServiceQueries.listProjectsByCurrentUser,
    create(ListProjectsByCurrentUserRequestSchema, { 
      orgId: organization?.id || '',
      withPermissions: ['update', 'delete'],
      nonInherited: true,
      withMemberCount 
    }),
    { enabled: !!organization?.id && !allProjects }
  );

  const refetch = useCallback(() => {
    if (allProjects) {
      refetchOrgProjects();
    } else {
      refetchUserProjects();
    }
  }, [allProjects, refetchOrgProjects, refetchUserProjects]);

  const projects = useMemo(() => {
    return allProjects 
      ? (orgProjectsData?.projects || [])
      : (userProjectsData?.projects || []);
  }, [allProjects, orgProjectsData?.projects, userProjectsData?.projects]);

  const accessPairs = useMemo(() => {
    return allProjects 
      ? [] // ListOrganizationProjectsResponse doesn't have accessPairs
      : (userProjectsData?.accessPairs || []);
  }, [allProjects, userProjectsData?.accessPairs]);

  const userAccessOnProject = useMemo(() => {
    return accessPairs.reduce((acc: Record<string, string[]>, p: ListProjectsByCurrentUserResponse_AccessPair) => {
      const { projectId, permissions } = p;
      acc[projectId] = permissions;
      return acc;
    }, {});
  }, [accessPairs]);

  const isLoading: boolean = allProjects ? isOrgProjectsLoading : isUserProjectsLoading;
  const error: unknown = allProjects ? orgProjectsError : userProjectsError;

  return {
    isFetching: isLoading,
    projects: projects,
    userAccessOnProject,
    refetch: refetch,
    error
  };
};
