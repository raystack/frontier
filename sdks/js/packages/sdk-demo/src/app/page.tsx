'use client';
import AuthContext from '@/contexts/auth';
import { Button, Flex } from '@raystack/apsara/v1';
import { useFrontier } from '@raystack/frontier/react';

import Link from 'next/link';
import { redirect } from 'next/navigation';
import { useContext, useEffect } from 'react';

import frontierClient from '@/api/frontier';

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);
  const { organizations } = useFrontier();
  useEffect(() => {
    if (!isAuthorized) {
      redirect('/login');
    }
  }, [isAuthorized]);

  async function logout() {
    const resp = await frontierClient?.frontierServiceAuthLogout();
    if (resp?.status === 200) {
      window.location.reload();
    }
  }

  return (
    <main>
      <Flex
        align="center"
        style={{ height: '100vh', width: '100vw' }}
        direction="column"
      >
        <Button
          variant="outline"
          color="neutral"
          size="small"
          data-test-id="[logout-button]"
          onClick={logout}
        >
          Logout
        </Button>
        <Flex direction="row" wrap="wrap">
          {organizations.map(org => (
            <Flex
              key={org.id}
              style={{
                padding: 'var(--rs-space-5)',
                border: '1px solid var(--rs-color-border-base-secondary)',
                margin: 'var(--rs-space-3)'
              }}
            >
              <Link href={`/organizations/${org.id}`} data-test-id={`[organization-link-${org.id}]`}>{org.title}</Link>
            </Flex>
          ))}
        </Flex>
      </Flex>
    </main>
  );
}
