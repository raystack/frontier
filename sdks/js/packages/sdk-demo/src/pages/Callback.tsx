import { useEffect, useCallback, Suspense } from 'react';
import frontierClient from '@/api/frontier';
import { Flex } from '@raystack/apsara/v1';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthRedirect from '@/hooks/useAuthRedirect';

function CallbackComponent() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const callFrontierCallback = useCallback(async () => {
    const state = searchParams?.get('state');
    const code = searchParams?.get('code');

    try {
      if (state && code) {
        const resp = await frontierClient?.frontierServiceAuthCallback({
          state,
          code
        });
        if (resp?.status === 200) {
          navigate('/');
        } else {
          throw new Error('Auth callback failed');
        }
      }
    } catch (err) {
      navigate('/login', { replace: true });
    }
  }, [navigate, searchParams]);

  useEffect(() => {
    callFrontierCallback();
  }, [callFrontierCallback]);

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
