'use client';

import { useNavigate } from '@tanstack/react-router';
import { BillingPage } from '~/react/views/billing';

export default function Billing() {
  const navigate = useNavigate({ from: '/billing' });

  return (
    <BillingPage
      onNavigateToPlans={() => navigate({ to: '/plans' })}
    />
  );
}
