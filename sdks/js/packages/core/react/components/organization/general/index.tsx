'use client';

import { Button, Flex, Separator, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS, formatPermissions } from '~/utils';
import { styles } from '../styles';
import { GeneralProfile } from './general.profile';
import { GeneralOrganization } from './general.workspace';

export default function GeneralSetting() {
  const [permisionValues, setPermisionValues] = useState([]);
  const [fetchingOrgPermissions, setFetchingOrgPermissions] = useState(true);

  const { client, activeOrganization: organization } = useFrontier();
  const isLoading = fetchingOrgPermissions;

  const PERMISSIONS_MAP = {
    Organization: `app/organization:${organization?.id}`
  };

  const permisions = [
    {
      permission: PERMISSIONS.GET,
      resource: PERMISSIONS_MAP.Organization
    },
    {
      permission: PERMISSIONS.DELETE,
      resource: PERMISSIONS_MAP.Organization
    }
  ];

  const fetchOrganizationPermissions = useCallback(async () => {
    try {
      const {
        // @ts-ignore
        data: { pairs }
      } = await client?.frontierServiceBatchCheckPermission({
        bodies: permisions
      });
      setPermisionValues(pairs);
    } catch (err) {
      console.error(err);
    } finally {
      setFetchingOrgPermissions(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client]);

  useEffect(() => {
    if (organization?.id) {
      fetchOrganizationPermissions();
    }
  }, [fetchOrganizationPermissions, organization?.id]);

  const organizationPermissions = useMemo(() => {
    if (permisionValues.length) {
      return formatPermissions(permisionValues);
    } else {
      return {};
    }
  }, [permisionValues]);

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
          permissionMap={PERMISSIONS_MAP}
          organizationPermissions={organizationPermissions}
          isLoading={isLoading}
        />
        <Separator />
        {organizationPermissions &&
        organizationPermissions[
          `${PERMISSIONS.DELETE}::${PERMISSIONS_MAP.Organization}`
        ] ? (
          <GeneralDeleteOrganization isLoading={isLoading} />
        ) : null}
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
    <>
      <Flex direction="column" gap="medium">
        <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
          If you want to permanently delete this organization and all of its
          data.
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
      <Separator />
    </>
  );
};


