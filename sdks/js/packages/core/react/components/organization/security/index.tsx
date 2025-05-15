'use client';

import { Box, Flex, Separator, Text } from '@raystack/apsara';
import { Switch } from '@raystack/apsara/v1';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Preference } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import type { SecurityCheckboxTypes } from './security.types';
import { styles } from '../styles';

export default function WorkspaceSecurity() {
  const [socialLogin, setSocialLogin] = useState<boolean>(false);
  const [mailLink, setMailLink] = useState<boolean>(false);

  const [preferences, setPreferences] = useState<V1Beta1Preference[]>([]);
  const { client, activeOrganization: organization } = useFrontier();

  const fetchOrganizationPreferences = useCallback(async () => {
    const {
      // @ts-ignore
      data: { preferences }
    } = await client?.frontierServiceListOrganizationPreferences(
      organization?.id as string
    );

    setPreferences(preferences);
  }, [client, organization?.id]);

  useEffect(() => {
    if (organization?.id) fetchOrganizationPreferences();
  }, [organization?.id, client, fetchOrganizationPreferences]);

  const preferencesMap = useMemo(() => {
    return preferences.reduce<Record<string, Record<string, string>>>(
      (map, el) => {
        // @ts-ignore
        map[el.name] = el;
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preferences]);

  const onValueChange = useCallback(
    async (key: string, checked: boolean) => {
      if (key === 'mail_link') setMailLink(checked);
      if (key === 'social_login') setSocialLogin(checked);
      await client?.frontierServiceCreateOrganizationPreferences(
        organization?.id as string,
        {
          bodies: [
            {
              name: key,
              value: `${checked}`
            }
          ]
        }
      );
    },
    [client, organization?.id]
  );

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
      <Flex style={styles.header}>
        <Text size={6}>Security</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <SecurityCheckbox
          label="Google"
          text="Allow logins through Google&#39;s single sign-on functionality"
          name="social_login"
          value={socialLogin}
          canUpdatePreference={canUpdatePreference}
          onValueChange={onValueChange}
        />
        <Separator></Separator>
        <SecurityCheckbox
          label="Email code"
          text="Allow password less logins through magic links or a code delivered
      over email."
          name="mail_link"
          value={mailLink}
          canUpdatePreference={canUpdatePreference}
          onValueChange={onValueChange}
        />
        <Separator></Separator>
      </Flex>
    </Flex>
  );
}

export const SecurityHeader = () => {
  return (
    <Box style={styles.container}>
      <Text size={10}>Security</Text>
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Manage your workspace security and how it’s members authenticate
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
      <Flex direction="column" gap="small">
        <Text size={6}>{label}</Text>
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
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
