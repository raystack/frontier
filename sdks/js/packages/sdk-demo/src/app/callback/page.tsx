'use client';

import { useEffect, useCallback, Suspense } from 'react';
import frontierClient from '@/api/frontier';
import { Flex } from '@raystack/apsara';
import { useSearchParams, useRouter } from 'next/navigation';
import useAuthRedirect from '@/hooks/useAuthRedirect';

function Callback() {
  const searchParams = useSearchParams();
  const router = useRouter();

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
          router.push('/');
        } else {
          throw new Error('Auth callback failed');
        }
      }
    } catch (err) {
      router.replace('/login');
    }
  }, [router, searchParams]);

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

export default function Page() {
  useAuthRedirect();

  return (
    <Suspense>
      <Callback />
    </Suspense>
  );
}
