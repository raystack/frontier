import AuthContext from '@/contexts/auth';
import { useNavigate } from 'react-router-dom';
import { useContext, useEffect } from 'react';

const useAuthRedirect = () => {
  const { isAuthorized } = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthorized) {
      navigate('/');
    }
  }, [isAuthorized, navigate]);
};

export default useAuthRedirect;
