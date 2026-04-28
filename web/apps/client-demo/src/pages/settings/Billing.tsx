import { BillingView } from '@raystack/frontier/client';
import { useNavigate, useParams } from 'react-router-dom';

export default function Billing() {
  const { orgId } = useParams<{ orgId: string }>();
  const navigate = useNavigate();
  return <BillingView onNavigateToPlans={() => navigate(`/${orgId}/settings/plans`)} />;
}
