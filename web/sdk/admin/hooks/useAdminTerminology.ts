import { useAdminConfig } from "../contexts/AdminConfigContext";
import { defaultTerminology } from "../utils/constants";

export interface TerminologyOptions {
  plural?: boolean;
  case?: "lower" | "upper" | "capital";
}

export interface TerminologyEntity {
  (options?: TerminologyOptions): string;
}

const applyCase = (
  text: string,
  caseType?: "lower" | "upper" | "capital"
): string => {
  switch (caseType) {
    case "lower":
      return text.toLowerCase();
    case "upper":
      return text.toUpperCase();
    case "capital":
      return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
    default:
      return text;
  }
};

const createEntity = (singular: string, plural: string): TerminologyEntity => {
  return ({
    plural: isPlural = false,
    case: caseType,
  }: TerminologyOptions = {}) => {
    const text = isPlural ? plural : singular;
    return applyCase(text, caseType);
  };
};

export const useAdminTerminology = () => {
  const config = useAdminConfig();
  const terminology = config.terminology || defaultTerminology;

  return {
    organization: createEntity(
      terminology.organization?.singular || defaultTerminology.organization.singular,
      terminology.organization?.plural || defaultTerminology.organization.plural
    ),
    project: createEntity(
      terminology.project?.singular || defaultTerminology.project.singular,
      terminology.project?.plural || defaultTerminology.project.plural
    ),
    team: createEntity(
      terminology.team?.singular || defaultTerminology.team.singular,
      terminology.team?.plural || defaultTerminology.team.plural
    ),
    member: createEntity(
      terminology.member?.singular || defaultTerminology.member.singular,
      terminology.member?.plural || defaultTerminology.member.plural
    ),
    user: createEntity(
      terminology.user?.singular || defaultTerminology.user.singular,
      terminology.user?.plural || defaultTerminology.user.plural
    ),
    appName: createEntity(
      terminology.appName || defaultTerminology.appName,
      terminology.appName || defaultTerminology.appName
    ),
  };
};
