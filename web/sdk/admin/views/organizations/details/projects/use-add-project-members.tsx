import { useCallback, useContext, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { toast } from "@raystack/apsara";
import { DEFAULT_ROLES } from "../../../../utils/constants";
import { useQuery, useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListProjectUsersRequestSchema, SetProjectMemberRoleRequestSchema } from "@raystack/proton/frontier";
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

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole,
  );

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !projectId) return;
      try {
        await setProjectMemberRole(
          create(SetProjectMemberRoleRequestSchema, {
            projectId,
            principalId: userId,
            principalType: "app/user",
            roleId: DEFAULT_ROLES.PROJECT_VIEWER,
          }),
        );
        toast.success(`${memberLabel} added`);
        await refetch();
        return projectMembers;
      } catch (error: unknown) {
        handleConnectError(error, {
          AlreadyExists: () => toast.error(`${memberLabel} already exists in this project`),
          PermissionDenied: () => toast.error("You don't have permission to perform this action"),
          InvalidArgument: (err) => toast.error('Invalid input', { description: err.message }),
          Default: (err) => toast.error('Something went wrong', { description: err.message }),
        });
      }
    },
    [projectId, setProjectMemberRole, refetch, projectMembers, memberLabel],
  );

  return {
    isLoading,
    eligibleMembers,
    fetchProjectMembers: refetch,
    setSearchQuery,
    addMember,
  };
}
