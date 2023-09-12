'use client';

import { Button, Flex, Separator, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { styles } from '../styles';
import { GeneralProfile } from './general.profile';
import { GeneralOrganization } from './general.workspace';

export default function GeneralSetting() {
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading: isLoading
  } = useFrontier();
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>General</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <GeneralProfile organization={organization} isLoading={isLoading} />
        <Separator />
        <GeneralOrganization
          organization={organization}
          isLoading={isLoading}
        />
        <Separator />
        <GeneralDeleteOrganization isLoading={isLoading} />
        <Separator />
      </Flex>
    </Flex>
  );
}

export const GeneralDeleteOrganization = ({
  isLoading
}: {
  isLoading?: boolean;
}) => {
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
        disabled={isLoading}
      >
        Delete organization
      </Button>
      <Outlet />
    </Flex>
  );
};
