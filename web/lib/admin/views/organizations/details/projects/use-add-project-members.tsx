import { useCallback, useContext, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { toast } from "@raystack/apsara";
import { DEFAULT_ROLES } from "../../../../utils/constants";
import { useQuery, useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListProjectUsersRequestSchema, CreatePolicyRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

interface useAddProjectMembersProps {
  projectId: string;
}

export function useAddProjectMembers({ projectId }: useAddProjectMembersProps) {
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

  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy,
  );

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !projectId) return;
      try {
        const principal = `app/user:${userId}`;
        const resource = `app/project:${projectId}`;
        await createPolicy(
          create(CreatePolicyRequestSchema, {
            body: {
              roleId: DEFAULT_ROLES.PROJECT_VIEWER,
              principal,
              resource,
            },
          }),
        );
        toast.success("member added");
        await refetch();
        return projectMembers;
      } catch (error: unknown) {
        console.error(error);
        toast.error("Something went wrong");
      }
    },
    [projectId, createPolicy, refetch, projectMembers],
  );

  return {
    isLoading,
    eligibleMembers,
    fetchProjectMembers: refetch,
    setSearchQuery,
    addMember,
  };
}
