import { useEffect, Suspense, useContext } from 'react';
import { Flex } from '@raystack/apsara';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthRedirect from '@/hooks/useAuthRedirect';
import { FrontierServiceQueries, useQuery } from '@raystack/frontier/hooks';
import AuthContext from '@/contexts/auth';

function CallbackComponent() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setIsAuthorized } = useContext(AuthContext);

  const state = searchParams.get('state') || '';
  const code = searchParams.get('code') || '';

  const { isSuccess, isError, error, isLoading } = useQuery(
    FrontierServiceQueries.authCallback,
    {
      state,
      code
    },
    { enabled: !!state && !!code }
  );

  useEffect(() => {
    if (!isLoading) {
      if (isSuccess) {
        navigate('/');
        setIsAuthorized(true);
      } else if (isError) {
        console.error('Auth callback failed:', error);
        setIsAuthorized(false);
        navigate('/login', { replace: true });
      }
    }
  }, [isSuccess, isError, navigate, error, isLoading, setIsAuthorized]);

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      Loading...
    </Flex>
  );
}

export default function Callback() {
  useAuthRedirect();

  return (
    <Suspense>
      <CallbackComponent />
    </Suspense>
  );
}
