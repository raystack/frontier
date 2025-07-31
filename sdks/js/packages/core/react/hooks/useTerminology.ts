import { useContext } from 'react';
import {
  CustomizationContext,
  defaultCustomization
} from '../contexts/CustomizationContext';

export interface TerminologyOptions {
  plural?: boolean;
  case?: 'lower' | 'upper' | 'capital';
}

export interface TerminologyEntity {
  (options?: TerminologyOptions): string;
}

const applyCase = (text: string, caseType?: 'lower' | 'upper' | 'capital'): string => {
  switch (caseType) {
    case 'lower':
      return text.toLowerCase();
    case 'upper':
      return text.toUpperCase();
    case 'capital':
      return text.charAt(0).toUpperCase() + text.slice(1).toLowerCase();
    default:
      return text;
  }
};

const createEntity = (singular: string, plural: string): TerminologyEntity => {
  return ({ plural: isPlural = false, case: caseType }: TerminologyOptions = {}) => {
    const text = isPlural ? plural : singular;
    return applyCase(text, caseType);
  };
};

export const useTerminology = () => {
  const context = useContext(CustomizationContext);
  const config = context || defaultCustomization;
  const { terminology } = config;

  return {
    organization: createEntity(
      terminology.organization!.singular,
      terminology.organization!.plural
    ),
    project: createEntity(
      terminology.project!.singular,
      terminology.project!.plural
    ),
    team: createEntity(terminology.team!.singular, terminology.team!.plural),
    member: createEntity(
      terminology.member!.singular,
      terminology.member!.plural
    ),
    user: createEntity(terminology.user!.singular, terminology.user!.plural),
    appName: createEntity(terminology.appName!, terminology.appName!)
  };
};
