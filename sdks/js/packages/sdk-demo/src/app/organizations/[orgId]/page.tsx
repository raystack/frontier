'use client';
import { Window, OrganizationProfile } from '@raystack/frontier/react';

export default function OrgPage({ params }: { params: { orgId: string } }) {
  const orgId = params.orgId;

  return orgId ? (
    <Window open={true} onOpenChange={() => {}}>
      <OrganizationProfile
        organizationId={orgId}
        showBilling={true}
        showTokens={true}
        showPreferences={true}
        showAPIKeys={true}
      />
    </Window>
  ) : null;
}
