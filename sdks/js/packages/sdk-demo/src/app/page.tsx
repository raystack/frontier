'use client';
import AuthContext from '@/contexts/auth';
import { redirect } from 'next/navigation';
import { useContext, useEffect } from 'react';

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);

  useEffect(() => {
    if (!isAuthorized) {
      redirect('/login');
    }
  }, [isAuthorized]);

  return (
    <main>
      <h1>Frontier SDK</h1>
    </main>
  );
}
