'use client';

import { Switch, Separator, Box, Text, Headline, Flex, toast } from '@raystack/apsara';
import { useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { useQuery, useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries, ListOrganizationPreferencesRequestSchema, CreateOrganizationPreferencesRequestSchema, type Preference } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import type { SecurityCheckboxTypes } from './security.types';
import { PageHeader } from '~/react/components/common/page-header';
import sharedStyles from '../styles.module.css';

export default function WorkspaceSecurity() {
  const [socialLogin, setSocialLogin] = useState<boolean>(false);
  const [mailLink, setMailLink] = useState<boolean>(false);

  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const {
    data: preferencesData,
    error: preferencesError
  } = useQuery(
    FrontierServiceQueries.listOrganizationPreferences,
    create(ListOrganizationPreferencesRequestSchema, {
      id: organization?.id || ''
    }),
    {
      enabled: !!organization?.id
    }
  );

  const preferences = useMemo(() => preferencesData?.preferences ?? [], [preferencesData]);

  useEffect(() => {
    if (preferencesError) {
      toast.error('Something went wrong', {
        description: preferencesError.message
      });
    }
  }, [preferencesError]);

  const { mutate: createOrganizationPreferences } = useMutation(
    FrontierServiceQueries.createOrganizationPreferences,
    {
      onSuccess: () => {
        toast.success('Preference updated successfully');
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listOrganizationPreferences,
            transport,
            input: create(ListOrganizationPreferencesRequestSchema, {
              id: organization?.id || ''
            }),
            cardinality: 'finite'
          })
        });
      },
      onError: (error: Error) => {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
  );

  const preferencesMap = useMemo(() => {
    return preferences.reduce<Record<string, Preference>>(
      (map, el) => {
        if (el.name) map[el.name] = el;
        return map;
      },
      {}
    );
  }, [preferences]);

  useEffect(() => {
    if (preferencesMap['social_login'])
      setSocialLogin(preferencesMap['social_login']?.value === 'true');

    if (preferencesMap['mail_link'])
      setMailLink(preferencesMap['mail_link']?.value === 'true');
  }, [preferencesMap]);

  const onValueChange = (key: string, checked: boolean) => {
    if (!organization?.id) return;
    
    if (key === 'mail_link') setMailLink(checked);
    if (key === 'social_login') setSocialLogin(checked);
    
    createOrganizationPreferences(
      create(CreateOrganizationPreferencesRequestSchema, {
        id: organization.id,
        bodies: [
          {
            name: key,
            value: `${checked}`
          }
        ]
      })
    );
  };

  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: `app/organization:${organization?.id}`
      }
    ],
    [organization?.id]
  );

  const { permissions } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const canUpdatePreference = shouldShowComponent(
    permissions,
    `${PERMISSIONS.UpdatePermission}::app/organization:${organization?.id}`
  );

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <PageHeader 
            title="Security" 
            description="Manage your workspace security and how it's members authenticate"
          />
        </Flex>
        <Flex direction="column" gap={9}>
        <SecurityCheckbox
          label="Google"
          text="Allow logins through Google's single sign-on functionality"
          name="social_login"
          value={socialLogin}
          canUpdatePreference={canUpdatePreference}
          onValueChange={onValueChange}
        />
        <Separator />
        <SecurityCheckbox
          label="Email code"
          text="Allow password less logins through magic links or a code delivered
      over email."
          name="mail_link"
          value={mailLink}
          canUpdatePreference={canUpdatePreference}
          onValueChange={onValueChange}
        />
        <Separator />
        </Flex>
      </Flex>
    </Flex>
  );
}

export const SecurityHeader = () => {
  return (
    <Box>
      <Headline size="t3">Security</Headline>
      <Text size="regular" variant="secondary">
        Manage your workspace security and how it&apos;s members authenticate
      </Text>
    </Box>
  );
};

export const SecurityCheckbox = ({
  label,
  text,
  name,
  value,
  onValueChange,
  canUpdatePreference
}: SecurityCheckboxTypes) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap={3}>
        <Text size="large">{label}</Text>
        <Text size="regular" variant="secondary">
          {text}
        </Text>
      </Flex>

      {canUpdatePreference ? (
        <Switch
          name={name}
          checked={value}
          onCheckedChange={(checked: boolean) => onValueChange(name, checked)}
        />
      ) : null}
    </Flex>
  );
};
