import { createContext, useContext } from "react";
import { Group } from "../types";

export type GroupContextProps = {
  groups: Group[];
};
const initialValues: GroupContextProps = {
  groups: [],
};
export const GroupContext = createContext<GroupContextProps>(initialValues);
GroupContext.displayName = "GroupContext ";

export function useGroupContext() {
  const context = useContext<GroupContextProps>(GroupContext);
  return context ? context : (initialValues as GroupContextProps);
}
