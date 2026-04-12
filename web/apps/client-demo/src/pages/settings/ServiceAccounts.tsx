import { useParams, useNavigate } from 'react-router-dom';
import { ServiceAccountsView } from '@raystack/frontier/react';

export default function ServiceAccounts() {
  const { orgId } = useParams<{ orgId: string }>();
  const navigate = useNavigate();

  return (
    <ServiceAccountsView
      onServiceAccountClick={id =>
        navigate(`/${orgId}/settings/service-accounts/${id}`)
      }
    />
  );
}
