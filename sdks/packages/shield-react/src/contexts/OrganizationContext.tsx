import { createContext, useContext } from "react";
import { Organization } from "../types/organization";

export type OrganizationContextProps = {
  organizations: Organization[];
};
const initialValues: OrganizationContextProps = {
  organizations: [],
};
export const OrganizationContext =
  createContext<OrganizationContextProps>(initialValues);
OrganizationContext.displayName = "OrganizationContext ";

export function useOrganizationContext() {
  const context = useContext<OrganizationContextProps>(OrganizationContext);
  return context ? context : (initialValues as OrganizationContextProps);
}
