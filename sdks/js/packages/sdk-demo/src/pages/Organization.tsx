import { Window, OrganizationProfile } from '@raystack/frontier/react';
import { useParams } from 'react-router-dom';

export default function Organization() {
  const { orgId } = useParams<{ orgId: string }>();

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
