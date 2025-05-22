import { createContext, useContext } from "react";
import { V1Beta1User } from "~/api/frontier";

const UserContext = createContext<V1Beta1User | undefined>(undefined);

export const UserProvider = UserContext.Provider;

export const useUser = () => {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error("useUser must be used within a UserProvider");
  }
  return context;
};
