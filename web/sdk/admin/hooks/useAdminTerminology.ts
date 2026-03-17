import { useMemo } from "react";
import { useAdminConfig } from "../contexts/AdminConfigContext";
import { defaultTerminology } from "../utils/constants";
import {
  createTerminologyMap,
  type TerminologyMap,
} from "../../shared/terminology";

// Re-export types so existing imports don't break
export type {
  TerminologyOptions,
  TerminologyEntity,
  TerminologyMap,
} from "../../shared/terminology";

export const useAdminTerminology = (): TerminologyMap => {
  const config = useAdminConfig();
  const terminology = config.terminology || defaultTerminology;

  return useMemo(
    () => createTerminologyMap(terminology, defaultTerminology),
    [terminology]
  );
};
