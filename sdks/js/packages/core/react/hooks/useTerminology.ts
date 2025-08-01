import { useCustomizationContext } from '../contexts/CustomizationContext';
import { EntityTerminologies } from '../../shared/types';

export interface TerminologyOptions {
  plural?: boolean;
  case?: 'lower' | 'upper' | 'capital';
}

export interface TerminologyEntity {
  (options?: TerminologyOptions): string;
}

const applyCase = (
  text: string,
  caseType?: 'lower' | 'upper' | 'capital'
): string => {
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
  return ({
    plural: isPlural = false,
    case: caseType
  }: TerminologyOptions = {}) => {
    const text = isPlural ? plural : singular;
    return applyCase(text, caseType);
  };
};

export const useTerminology = () => {
  const context = useCustomizationContext();
  const { terminology } = context;

  return Object.entries(terminology)
    .filter(([key]) => key !== 'appName')
    .reduce((acc, [key, value]) => {
      const entity = value as EntityTerminologies;
      acc[key] = createEntity(entity.singular, entity.plural);
      return acc;
    }, {
      appName: createEntity(terminology.appName!, terminology.appName!)
    } as Record<string, TerminologyEntity>);
};
