import { useCallback, useContext, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { toastManager } from "@raystack/apsara-v1";
import { DEFAULT_ROLES, SCOPES } from "~/admin/utils/constants";
import { useQuery, useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListProjectUsersRequestSchema, ListRolesRequestSchema, SetProjectMemberRoleRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { handleConnectError } from "~/utils/error";
import { useTerminology } from "../../../../hooks/useTerminology";

interface useAddProjectMembersProps {
  projectId: string;
}

export function useAddProjectMembers({ projectId }: useAddProjectMembersProps) {
  const t = useTerminology();
  const memberLabel = t.member({ case: "capital" });
  const { orgMembersMap } = useContext(OrganizationContext);
  const [searchQuery, setSearchQuery] = useState<string>("");

  const { data: projectMembers, isLoading, refetch } = useQuery(
    FrontierServiceQueries.listProjectUsers,
    create(ListProjectUsersRequestSchema, { id: projectId }),
    {
      enabled: !!projectId,
      select: (data) => data?.users || [],
    }
  );

  const nonMembers = useMemo(() => {
    if (!projectMembers) return [];
    const memberSet = new Set(projectMembers.map((member) => member.id));
    return Object.values(orgMembersMap)
      .filter((member) => !memberSet.has(member.id))
      .sort((a, b) => {
        const aName = a.title || a.email || "";
        const bName = b.title || b.email || "";
        return aName < bName ? -1 : 1;
      });
  }, [projectMembers, orgMembersMap]);

  const eligibleMembers = useMemo(() => {
    return searchQuery
      ? nonMembers.filter((member) => {
          const name = member.title || member.email || "";
          return name.toLowerCase().includes(searchQuery.toLowerCase());
        })
      : nonMembers;
  }, [nonMembers, searchQuery]);

  const { data: rolesData } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: "enabled",
      scopes: [SCOPES.PROJECT],
    }),
    { enabled: !!projectId }
  );

  const viewerRoleId = useMemo(
    () => rolesData?.roles?.find((r) => r.name === DEFAULT_ROLES.PROJECT_VIEWER)?.id ?? "",
    [rolesData],
  );

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole,
  );

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !projectId || !viewerRoleId) return;
      try {
        await setProjectMemberRole(
          create(SetProjectMemberRoleRequestSchema, {
            projectId,
            principalId: userId,
            principalType: SCOPES.USER,
            roleId: viewerRoleId,
          }),
        );
        toastManager.add({ title: `${memberLabel} added`, type: "success" });
        await refetch();
        return projectMembers;
      } catch (error: unknown) {
        handleConnectError(error, {
          AlreadyExists: () => toastManager.add({ title: `${memberLabel} already exists in this project`, type: "error" }),
          PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: "error" }),
          InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: "error" }),
          Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: "error" }),
        });
      }
    },
    [projectId, setProjectMemberRole, refetch, projectMembers, memberLabel, viewerRoleId],
  );

  return {
    isLoading,
    eligibleMembers,
    fetchProjectMembers: refetch,
    setSearchQuery,
    addMember,
  };
}
