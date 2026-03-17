import { EntityTerminologies } from "./types";

export interface TerminologyOptions {
  plural?: boolean;
  case?: "lower" | "upper" | "capital";
}

export interface TerminologyEntity {
  (options?: TerminologyOptions): string;
}

export const applyCase = (
  text: string,
  caseType?: "lower" | "upper" | "capital"
): string => {
  switch (caseType) {
    case "lower":
      return text.toLowerCase();
    case "upper":
      return text.toUpperCase();
    case "capital":
      return text.charAt(0).toUpperCase() + text.slice(1);
    default:
      return text;
  }
};

export const createEntity = (
  singular: string,
  plural: string
): TerminologyEntity => {
  return ({
    plural: isPlural = false,
    case: caseType,
  }: TerminologyOptions = {}) => {
    const text = isPlural ? plural : singular;
    return applyCase(text, caseType);
  };
};

export interface TerminologyConfig {
  organization?: EntityTerminologies;
  project?: EntityTerminologies;
  team?: EntityTerminologies;
  member?: EntityTerminologies;
  user?: EntityTerminologies;
  appName?: string;
}

const DEFAULT_TERMINOLOGY: Required<TerminologyConfig> = {
  organization: { singular: "Organization", plural: "Organizations" },
  project: { singular: "Project", plural: "Projects" },
  team: { singular: "Team", plural: "Teams" },
  member: { singular: "Member", plural: "Members" },
  user: { singular: "User", plural: "Users" },
  appName: "Frontier",
};

export interface TerminologyMap {
  organization: TerminologyEntity;
  project: TerminologyEntity;
  team: TerminologyEntity;
  member: TerminologyEntity;
  user: TerminologyEntity;
  appName: TerminologyEntity;
}

export function createTerminologyMap(
  terminology?: TerminologyConfig,
  defaults: Required<TerminologyConfig> = DEFAULT_TERMINOLOGY
): TerminologyMap {
  const t = terminology || defaults;
  return {
    organization: createEntity(
      t.organization?.singular || defaults.organization.singular,
      t.organization?.plural || defaults.organization.plural
    ),
    project: createEntity(
      t.project?.singular || defaults.project.singular,
      t.project?.plural || defaults.project.plural
    ),
    team: createEntity(
      t.team?.singular || defaults.team.singular,
      t.team?.plural || defaults.team.plural
    ),
    member: createEntity(
      t.member?.singular || defaults.member.singular,
      t.member?.plural || defaults.member.plural
    ),
    user: createEntity(
      t.user?.singular || defaults.user.singular,
      t.user?.plural || defaults.user.plural
    ),
    appName: createEntity(
      t.appName || defaults.appName,
      t.appName || defaults.appName
    ),
  };
}
