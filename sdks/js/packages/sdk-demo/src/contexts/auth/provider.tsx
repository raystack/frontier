'use client';

import React, { PropsWithChildren, useState } from 'react';
import AuthContext from '.';

const AuthContextProvider: React.FC<PropsWithChildren> = ({ children }) => {
  const [isAuthorized, setIsAuthorized] = useState(false);

  return (
    <AuthContext.Provider value={{ isAuthorized, setIsAuthorized }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
