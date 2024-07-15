import AuthContext from '@/contexts/auth';
import { redirect } from 'next/navigation';
import { useContext, useEffect } from 'react';

const useAuthRedirect = () => {
  const { isAuthorized } = useContext(AuthContext);

  useEffect(() => {
    if (isAuthorized) {
      redirect('/');
    }
  }, [isAuthorized]);
};

export default useAuthRedirect;
