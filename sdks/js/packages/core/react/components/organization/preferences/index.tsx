'use client';

import { GearIcon, MoonIcon, SunIcon } from '@radix-ui/react-icons';
import { Flex, useTheme, Text, Separator, Box, Skeleton, Image, Headline } from '@raystack/apsara/v1';
import { Select } from '@raystack/apsara';
import bell from '~/react/assets/bell.svg';
import bellSlash from '~/react/assets/bell-slash.svg';
import { styles } from '../styles';
import { PreferencesSelectionTypes } from './preferences.types';
import { usePreferences } from '~/react/hooks/usePreferences';
import { PREFERENCE_OPTIONS } from '~/react/utils/constants';

const themeOptions = [
  {
    title: (
      <Flex align="center" gap="small">
        <SunIcon /> Light
      </Flex>
    ),
    value: 'light'
  },
  {
    title: (
      <Flex align="center" gap="small">
        <MoonIcon /> Dark
      </Flex>
    ),
    value: 'dark'
  },
  {
    title: (
      <Flex align="center" gap="small">
        <GearIcon /> System
      </Flex>
    ),
    value: 'system'
  }
];
const newsletterOptions = [
  {
    title: (
      <Flex align="center" gap="small">
        <Image alt="close" width={16} height={16} src={bell as unknown as string} /> Subscribed
      </Flex>
    ),
    value: 'true'
  },
  {
    title: (
      <Flex align="center" gap="small">
        <Image alt="close" width={16} height={16} src={bellSlash as unknown as string} />{' '}
        Unsubscribed
      </Flex>
    ),
    value: 'false'
  }
];

export default function UserPreferences() {
  const { theme, setTheme } = useTheme();
  const { preferences, isLoading, isFetching, updatePreferences } =
    usePreferences({});

  const newsletterValue =
    preferences?.[PREFERENCE_OPTIONS.NEWSLETTER]?.value ?? 'false';

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size="large">Preferences</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <PreferencesSelection
          label="Theme"
          text="Customise your interface color scheme."
          name="theme"
          defaultValue={theme}
          values={themeOptions}
          onSelection={value => setTheme(value)}
        />
        <Separator />
        <PreferencesSelection
          label="Updates, News & Events"
          text="Stay informed on new features, improvements, and key updates."
          name={PREFERENCE_OPTIONS.NEWSLETTER}
          defaultValue={newsletterValue}
          values={newsletterOptions}
          isLoading={isFetching}
          disabled={isLoading}
          onSelection={value => {
            updatePreferences([{ name: PREFERENCE_OPTIONS.NEWSLETTER, value }]);
          }}
        />
        <Separator />
      </Flex>
    </Flex>
  );
}

export const PreferencesHeader = () => {
  return (
    <Box style={styles.container}>
      <Headline size="t3">Preferences</Headline>
      <Text size="regular" variant="secondary">
        Manage your workspace security and how it&apos;s members authenticate
      </Text>
    </Box>
  );
};

export const PreferencesSelection = ({
  label,
  text,
  name,
  values,
  defaultValue,
  isLoading = false,
  disabled = false,
  onSelection
}: PreferencesSelectionTypes) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size="regular">{label}</Text>
        <Text size="small" variant="secondary">
          {text}
        </Text>
      </Flex>
      {isLoading ? (
        <Skeleton width={120} height={32} />
      ) : (
        <Select
          onValueChange={onSelection}
          defaultValue={defaultValue}
          disabled={disabled}
          name={name}
        >
          <Select.Trigger>
            <Select.Value placeholder={label} />
          </Select.Trigger>
          <Select.Content style={{ minWidth: '120px' }}>
            <Select.Group>
              {values.map(v => (
                <Select.Item key={v.value} value={v.value}>
                  {v.title}
                </Select.Item>
              ))}
            </Select.Group>
          </Select.Content>
        </Select>
      )}
    </Flex>
  );
};
