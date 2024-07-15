import { createContext } from 'react';

interface AuthContextType {
  isAuthorized: boolean;
  setIsAuthorized: (v: boolean) => void;
}

const AuthContext = createContext<AuthContextType>({
  isAuthorized: false,
  setIsAuthorized: () => {}
});

export default AuthContext;
