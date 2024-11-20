import { Flex, Text, EmptyState, Button } from '@raystack/apsara/v1';
import styles from './styles.module.css';
import keyIcon from '~/react/assets/key.svg';
import { Image } from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_API_PLATFORM_APP_NAME } from '~/react/utils/constants';
import { FrontierClientAPIPlatformOptions } from '~/shared/types';
import { useMemo } from 'react';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';

const NoServiceAccounts = ({
  config
}: {
  config?: FrontierClientAPIPlatformOptions;
}) => {
  const appName = config?.appName || DEFAULT_API_PLATFORM_APP_NAME;
  return (
    <EmptyState
      icon={
        <Image
          // @ts-ignore
          src={keyIcon}
          alt="keyIcon"
        />
      }
      heading="No service account"
      subHeading={`Create a new account to use the APIs of ${appName}`}
      primaryAction={
        <Button
          data-test-id="frontier-sdk-new-service-account-btn"
          variant="secondary"
        >
          Create new service account
        </Button>
      }
    />
  );
};

const NoAccess = () => {
  return (
    <EmptyState
      icon={<ExclamationTriangleIcon />}
      heading="Restricted Access"
      subHeading={`Admin access required, please reach out to your admin incase you want to generate a key.`}
    />
  );
};

const useAccess = (orgId?: string) => {
  const resource = `app/organization:${orgId}`;
  const listOfPermissionsToCheck = useMemo(() => {
    return [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: resource
      }
    ];
  }, [resource]);

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!orgId
  );

  const canUpdateWorkspace = useMemo(() => {
    return shouldShowComponent(
      permissions,
      `${PERMISSIONS.UpdatePermission}::${resource}`
    );
  }, [permissions, resource]);

  return {
    isPermissionsFetching,
    canUpdateWorkspace
  };
};

export default function ApiKeys() {
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    config
  } = useFrontier();

  const { isPermissionsFetching, canUpdateWorkspace } = useAccess(
    organization?.id
  );

  // TODO: show skeleton loader for Keys List
  const isLoading = isActiveOrganizationLoading || isPermissionsFetching;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size={6}>API</Text>
      </Flex>
      <Flex justify="center" align="center" className={styles.content}>
        {canUpdateWorkspace ? (
          <NoServiceAccounts config={config.apiPlatform} />
        ) : (
          <NoAccess />
        )}
      </Flex>
    </Flex>
  );
}
