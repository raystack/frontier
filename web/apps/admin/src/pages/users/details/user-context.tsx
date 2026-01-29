import { User } from "@raystack/proton/frontier";
import { createContext, useContext } from "react";

interface UserContextType {
  user: User | undefined;
  reset?: () => void;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export const UserProvider = UserContext.Provider;

export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error("useUser must be used within a UserProvider");
  }
  return context;
};
