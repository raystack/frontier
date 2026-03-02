'use client';

import { useNavigate, useParams, useLocation } from '@tanstack/react-router';
import { ServiceUserDetailPage } from '~/react/views/api-keys';

export default function ServiceUserPage() {
  const { id } = useParams({ from: '/api-keys/$id' });
  const navigate = useNavigate({ from: '/api-keys/$id' });
  const location = useLocation();

  return (
    <ServiceUserDetailPage
      serviceUserId={id}
      onBack={() => navigate({ to: '/api-keys' })}
      enableTokensFetch={location.state?.enableServiceUserTokensListFetch}
    />
  );
}
