'use client';

import { useNavigate } from '@tanstack/react-router';
import { ApiKeysListPage } from '~/react/views/api-keys';

export default function APIKeys() {
  const navigate = useNavigate({ from: '/api-keys' });

  return (
    <ApiKeysListPage
      onServiceAccountClick={(id) =>
        navigate({
          to: '/api-keys/$id',
          params: { id }
        })
      }
    />
  );
}
