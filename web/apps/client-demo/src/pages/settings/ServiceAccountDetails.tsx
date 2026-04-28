import { useParams, useNavigate } from 'react-router-dom';
import { ServiceAccountDetailsView } from '@raystack/frontier/client';

export default function ServiceAccountDetails() {
  const { orgId, serviceAccountId } = useParams<{
    orgId: string;
    serviceAccountId: string;
  }>();
  const navigate = useNavigate();

  if (!serviceAccountId) return null;

  return (
    <ServiceAccountDetailsView
      serviceAccountId={serviceAccountId}
      onNavigateToServiceAccounts={() =>
        navigate(`/${orgId}/settings/service-accounts`)
      }
      onDeleteSuccess={() =>
        navigate(`/${orgId}/settings/service-accounts`)
      }
    />
  );
}
