import { useEffect, useCallback } from 'react';
import { useFrontier } from '@raystack/frontier/react';
import { Flex } from '@raystack/apsara';
import { redirect, useSearchParams } from 'next/navigation';
import useAuthRedirect from '@/hooks/useAuthRedirect';

export default function CallbackPage() {
  useAuthRedirect();

  const { client } = useFrontier();

  const searchParams = useSearchParams();

  const callFrontierCallback = useCallback(async () => {
    const state = searchParams.get('state');
    const code = searchParams.get('code');

    try {
      if (state && code) {
        const resp = await client?.frontierServiceAuthCallback({ state, code });
        if (resp?.status === 200) {
          redirect('/');
        } else {
          throw new Error('Auth callback failed');
        }
      }
    } catch (err) {
      redirect('/login');
    }
  }, [client, searchParams]);

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
