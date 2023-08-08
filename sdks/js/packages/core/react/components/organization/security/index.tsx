'use client';

import { Box, Flex, Separator, Switch, Text } from '@raystack/apsara';
import { styles } from '../styles';
import type { SecurityCheckboxTypes } from './security.types';

export default function WorkspaceSecurity() {
  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Security</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <SecurityCheckbox
          label="Google"
          text="Allow logins through Google&#39;s single sign-on functionality"
          name="google"
        />
        <Separator></Separator>
        <SecurityCheckbox
          label="Email code"
          text="Allow password less logins through magic links or a code delivered
      over email."
          name="email"
        />
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
  name
}: SecurityCheckboxTypes) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>{label}</Text>
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          {text}
        </Text>
      </Flex>
      {/* @ts-ignore */}
      <Switch name={name} />
    </Flex>
  );
};
