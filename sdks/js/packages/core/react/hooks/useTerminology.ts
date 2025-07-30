import { useContext } from 'react';
import {
  CustomizationContext,
  defaultCustomization
} from '../contexts/CustomizationContext';

export interface TerminologyOptions {
  plural?: boolean;
}

export interface TerminologyEntity {
  (options?: TerminologyOptions): string;
}

const createEntity = (singular: string, plural: string): TerminologyEntity => {
  return ({ plural: isPlural = false }: TerminologyOptions = {}) => {
    return isPlural ? plural : singular;
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
    appName: terminology.appName!
  };
};
