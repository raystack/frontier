import { useEffect, useCallback, Suspense } from 'react';
import { Flex } from '@raystack/apsara/v1';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthRedirect from '@/hooks/useAuthRedirect';
import { useMutation, FrontierServiceQueries } from '@raystack/frontier/hooks';

function CallbackComponent() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const { mutateAsync: authCallback } = useMutation(
    FrontierServiceQueries.authCallback
  );
  const state = searchParams?.get('state');
  const code = searchParams?.get('code');

  const callFrontierCallback = useCallback(
    async (state: string, code: string) => {
      try {
        await authCallback({
          state,
          code
        });
        navigate('/');
      } catch (err) {
        console.error('Auth callback failed:', err);
        navigate('/login', { replace: true });
      }
    },
    [authCallback, navigate]
  );

  useEffect(() => {
    if (state && code) {
      callFrontierCallback(state, code);
    }
  }, [state, code, callFrontierCallback]);

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
