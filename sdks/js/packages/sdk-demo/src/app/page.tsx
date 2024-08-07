'use client';
import AuthContext from '@/contexts/auth';
import { Flex } from '@raystack/apsara';
import { useFrontier } from '@raystack/frontier/react';
import Link from 'next/link';
import { redirect } from 'next/navigation';
import { useContext, useEffect } from 'react';

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);
  const { organizations } = useFrontier();
  useEffect(() => {
    if (!isAuthorized) {
      redirect('/login');
    }
  }, [isAuthorized]);

  return (
    <main>
      <Flex
        justify="center"
        align="center"
        style={{ height: '100vh', width: '100vw' }}
      >
        <Flex direction="column">
          {organizations.map(org => (
            <Flex
              key={org.id}
              style={{
                padding: '16px',
                border: '1px solid var(--border-base)',
                width: '100%',
                margin: '8px'
              }}
            >
              <Link href={`/organizations/${org.id}`}>{org.title}</Link>
            </Flex>
          ))}
        </Flex>
      </Flex>
    </main>
  );
}
