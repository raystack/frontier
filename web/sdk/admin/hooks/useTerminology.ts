// Re-export the shared useTerminology hook so existing admin view imports keep working.
// New code should import useTerminology directly.
export {
  useTerminology,
  useTerminology as useAdminTerminology,
  type TerminologyOptions,
  type TerminologyEntity,
  type TerminologyMap,
} from "../../shared/terminology";
