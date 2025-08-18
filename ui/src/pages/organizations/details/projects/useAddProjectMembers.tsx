import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import type {
  V1Beta1CreatePolicyForProjectBody,
  V1Beta1User,
} from "~/api/frontier";
import { OrganizationContext } from "../contexts/organization-context";
import { api } from "~/api";
import { toast } from "@raystack/apsara";
import { DEFAULT_ROLES } from "~/utils/constants";

interface useAddProjectMembersProps {
  projectId: string;
}

export function useAddProjectMembers({ projectId }: useAddProjectMembersProps) {
  const { orgMembersMap } = useContext(OrganizationContext);
  const [isLoading, setIsLoading] = useState(false);
  const [nonMembers, setNonMembers] = useState<V1Beta1User[]>([]);
  const [searchQuery, setSearchQuery] = useState<string>("");

  const fetchProjectMembers = useCallback(async () => {
    if (!projectId) {
      return;
    }
    try {
      setIsLoading(true);
      const resp = await api.frontierServiceListProjectUsers(projectId);
      const members = resp.data.users || [];
      const memberSet = new Set(members.map((member) => member.id));
      const projectNonMembers = Object.values(orgMembersMap)
        .filter((member) => !memberSet.has(member.id))
        .sort((a, b) => {
          const aName = a.title || a.email || "";
          const bName = b.title || b.email || "";
          return aName < bName ? -1 : 1;
        });
      setNonMembers(projectNonMembers);
      return members;
    } catch (error) {
      console.error("Error fetching project members:", error);
    } finally {
      setIsLoading(false);
    }
  }, [orgMembersMap, projectId]);

  const eligibleMembers = useMemo(() => {
    return searchQuery
      ? nonMembers.filter((member) => {
          const name = member.title || member.email || "";
          return name.toLowerCase().includes(searchQuery.toLowerCase());
        })
      : nonMembers;
  }, [nonMembers, searchQuery]);

  useEffect(() => {
    fetchProjectMembers();
  }, [fetchProjectMembers]);

  const addMember = useCallback(
    async (userId: string) => {
      if (!userId || !projectId) return;
      try {
        const principal = `app/user:${userId}`;
        const policy: V1Beta1CreatePolicyForProjectBody = {
          role_id: DEFAULT_ROLES.PROJECT_VIEWER,
          principal,
        };
        await api?.frontierServiceCreatePolicyForProject(projectId, policy);
        toast.success("member added");
        return fetchProjectMembers();
      } catch (error: unknown) {
        console.error(error);
        toast.error("Something went wrong");
      }
    },
    [projectId, fetchProjectMembers],
  );

  return {
    isLoading,
    eligibleMembers,
    fetchProjectMembers,
    setSearchQuery,
    addMember,
  };
}
