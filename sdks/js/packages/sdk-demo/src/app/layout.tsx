'use client';
import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { customFetch } from '@/utils/custom-fetch';
import { FrontierProvider } from '@raystack/frontier/react';
import type React from 'react';

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <FrontierProvider config={config} customFetch={customFetch}>
          <AuthContextProvider>{children}</AuthContextProvider>
        </FrontierProvider>
      </body>
    </html>
  );
}
