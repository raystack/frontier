import { useAdminConfig } from "../contexts/AdminConfigContext";
import { defaultTerminology } from "../utils/constants";

/**
 * Converts a terminology plural label into a URL-safe slug.
 * e.g. "Organizations" → "organizations", "Workspaces" → "workspaces"
 */
function toSlug(text: string): string {
  return text
    .toLowerCase()
    .replace(/\s+/g, "-")
    .replace(/[^a-z0-9-]/g, "");
}

export interface AdminPaths {
  /** URL slug for the organization entity, e.g. "organizations" or "workspaces" */
  organizations: string;
  /** URL slug for the user entity, e.g. "users" or "people" */
  users: string;
  /** URL slug for the project entity, e.g. "projects" or "repos" */
  projects: string;
  /** URL slug for the member entity, e.g. "members" or "participants" */
  members: string;
  /** URL slug for the team entity, e.g. "teams" or "groups" */
  teams: string;
}

export const useAdminPaths = (): AdminPaths => {
  const config = useAdminConfig();
  const terminology = config.terminology || defaultTerminology;

  return {
    organizations: toSlug(
      terminology.organization?.plural || defaultTerminology.organization.plural
    ),
    users: toSlug(
      terminology.user?.plural || defaultTerminology.user.plural
    ),
    projects: toSlug(
      terminology.project?.plural || defaultTerminology.project.plural
    ),
    members: toSlug(
      terminology.member?.plural || defaultTerminology.member.plural
    ),
    teams: toSlug(
      terminology.team?.plural || defaultTerminology.team.plural
    ),
  };
};
