'use client';

import { Button, Flex, Separator, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { styles } from '../styles';
import { GeneralProfile } from './general.profile';
import { GeneralOrganization } from './general.workspace';
import { useFrontier } from '~/react/contexts/FrontierContext';

export default function GeneralSetting() {
  const { activeOrganization: organization } = useFrontier();
  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>General</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <GeneralProfile organization={organization} />
        <Separator />
        <GeneralOrganization organization={organization} />
        <Separator />
        <GeneralDeleteOrganization />
        <Separator />
      </Flex>
    </Flex>
  );
}

export const GeneralDeleteOrganization = () => {
  const navigate = useNavigate({ from: '/' });
  return (
    <Flex direction="column" gap="medium">
      <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
        If you want to permanently delete this organization and all of its data.
      </Text>

      <Button
        variant="danger"
        type="submit"
        size="medium"
        onClick={() => navigate({ to: '/delete' })}
      >
        Delete organization
      </Button>
      <Outlet />
    </Flex>
  );
};
