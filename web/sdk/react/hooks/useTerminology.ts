import { useCustomizationContext } from '../contexts/CustomizationContext';
import {
  createTerminologyMap,
  type TerminologyMap,
} from '../../shared/terminology';

// Re-export types so existing imports don't break
export type {
  TerminologyOptions,
  TerminologyEntity,
  TerminologyMap,
} from '../../shared/terminology';

export const useTerminology = (): TerminologyMap => {
  const context = useCustomizationContext();
  const { terminology } = context;

  return createTerminologyMap(terminology);
};
