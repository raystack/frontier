import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { FrontierProvider } from '@raystack/frontier/react';
import type { Metadata } from 'next';
import React from 'react';

export const metadata: Metadata = {
  title: 'Frontier SDK',
  description: 'Frontier SDK'
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <FrontierProvider config={config}>
          <AuthContextProvider>{children}</AuthContextProvider>
        </FrontierProvider>
      </body>
    </html>
  );
}
