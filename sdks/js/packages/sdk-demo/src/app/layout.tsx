'use client';
import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { customFetch } from '@/utils/custom-fetch';
import {
  FrontierProvider,
  TranslationResources
} from '@raystack/frontier/react';
import type React from 'react';

const translations: TranslationResources = {
  en: {
    apiPlatform: {
      appName: 'Frontier Demo'
    },
    billing: {
      plan_change: {}
    },
    terminology: {
      organization: 'Workspace'
    }
  }
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <FrontierProvider
          config={config}
          customFetch={customFetch}
          translations={translations}
        >
          <AuthContextProvider>{children}</AuthContextProvider>
        </FrontierProvider>
      </body>
    </html>
  );
}
