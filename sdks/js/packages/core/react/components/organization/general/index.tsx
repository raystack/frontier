'use client';

import { Button, Flex, Separator, Text } from '@raystack/apsara';
import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization } from '~/src';
import { styles } from '../styles';
import { GeneralProfile } from './general.profile';
import { GeneralOrganization } from './general.workspace';

type GeneralSettingProps = {
  organization?: V1Beta1Organization;
};

export default function GeneralSetting({ organization }: GeneralSettingProps) {
  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>General</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <GeneralProfile />
        <Separator></Separator>
        <GeneralOrganization organization={organization} />
        <Separator></Separator>
        <GeneralDeleteOrganization organization={organization} />
      </Flex>
    </Flex>
  );
}

export const GeneralDeleteOrganization = ({
  organization
}: GeneralSettingProps) => {
  const { client } = useFrontier();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const organizationId = organization?.id;

  const onDeleteOrganization = useCallback(async () => {
    if (!organizationId) return;
    try {
      setIsSubmitting(true);
      await client?.frontierServiceDeleteOrganization(organizationId);
      // @ts-ignore
      window.location = `${window.location.origin}`;
    } catch ({ error }: any) {
      console.log(error);
      toast.error('Something went wrong', {
        description: `${error.message}`
      });
    }
  }, [client, organizationId]);

  return (
    <Flex direction="column" gap="medium">
      <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
        If you want to permanently delete this organization and all of its data.
      </Text>

      <Button
        variant="danger"
        type="submit"
        size="medium"
        onClick={onDeleteOrganization}
      >
        Delete {organization?.name}
      </Button>
    </Flex>
  );
};
