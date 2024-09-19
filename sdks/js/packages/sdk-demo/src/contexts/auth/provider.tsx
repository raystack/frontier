'use client';

import React, { PropsWithChildren, useEffect, useState } from 'react';
import AuthContext from '.';
import { useFrontier } from '@raystack/frontier/react';

const AuthContextProvider: React.FC<PropsWithChildren<{}>> = ({ children }) => {
  const [isAuthorized, setIsAuthorized] = useState(false);

  const { user, isUserLoading } = useFrontier();

  useEffect(() => {
    if (user?.id) {
      setIsAuthorized(true);
    } else if (!user?.id && !isUserLoading) {
      setIsAuthorized(false);
    }
  }, [user?.id, isUserLoading]);

  return (
    <AuthContext.Provider value={{ isAuthorized, setIsAuthorized }}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContextProvider;
