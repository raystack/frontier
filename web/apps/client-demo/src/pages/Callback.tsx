import { useEffect, Suspense, useContext } from 'react';
import { Flex } from '@raystack/apsara';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthRedirect from '@/hooks/useAuthRedirect';
import { handoffAuthRejection } from '@/hooks/useAuthRejection';
import {
  Code,
  ConnectError,
  FrontierServiceQueries,
  useQuery
} from '@raystack/frontier/hooks';
import { describeAuthError } from '@raystack/frontier/client';
import AuthContext from '@/contexts/auth';

const REJECTION_ROUTE: Partial<Record<Code, string>> = {
  [Code.AlreadyExists]: '/signup',
  [Code.FailedPrecondition]: '/signup',
  [Code.NotFound]: '/login'
};
const DEFAULT_REJECTION_ROUTE = '/login';

function CallbackComponent() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setIsAuthorized } = useContext(AuthContext);

  const state = searchParams.get('state') || '';
  const code = searchParams.get('code') || '';

  const { isSuccess, isError, error } = useQuery(
    FrontierServiceQueries.authCallback,
    {
      state,
      code
    },
    { enabled: !!state && !!code, retry: false }
  );

  useEffect(() => {
    if (isSuccess) {
      setIsAuthorized(true);
      navigate('/', { replace: true });
    } else if (isError) {
      setIsAuthorized(false);
      handoffAuthRejection(describeAuthError(error));
      const route = REJECTION_ROUTE[ConnectError.from(error).code];
      navigate(route ?? DEFAULT_REJECTION_ROUTE, { replace: true });
    }
  }, [isSuccess, isError, error, navigate, setIsAuthorized]);

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
