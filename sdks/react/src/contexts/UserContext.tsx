import { createContext, useContext } from "react";
import { User } from "../types";

type UserContextProps = {
  user: User | null;
};
const initialValues: UserContextProps = {
  user: null,
};
export const UserContext = createContext<UserContextProps>(initialValues);
UserContext.displayName = "UserContext ";

export function useUserContext() {
  const context = useContext(UserContext);
  return context ? context : (initialValues as UserContextProps);
}
