'use client';

import { GearIcon, MoonIcon, SunIcon } from '@radix-ui/react-icons';
import {
  Box,
  Flex,
  Image,
  Select,
  Separator,
  Text,
  useTheme
} from '@raystack/apsara';
import close from '~/react/assets/close.svg';
import open from '~/react/assets/open.svg';
import { styles } from '../styles';
import { PreferencesSelectionTypes } from './preferences.types';

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
        <MoonIcon /> dark
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
const sidebarOptions = [
  {
    title: (
      <Flex align="center" gap="small">
        {/* @ts-ignore */}
        <Image alt="open" width={16} height={16} src={open} /> Open
      </Flex>
    ),
    value: 'open'
  },
  {
    title: (
      <Flex align="center" gap="small">
        {/* @ts-ignore */}
        <Image alt="close" width={16} height={16} src={close} /> Collapsed
      </Flex>
    ),
    value: 'collapsed'
  }
];

export default function UserPreferences() {
  const { themes, theme, setTheme } = useTheme();

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Preferences</Text>
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
        <Separator></Separator>
        {/* <PreferencesSelection
          label="Sidebar"
          text="Select the default state of product sidebar."
          name="sidebar"
          defaultValue="open"
          values={sidebarOptions}
          onSelection={value => console.log(value)}
        />
        <Separator></Separator> */}
      </Flex>
    </Flex>
  );
}

export const PreferencesHeader = () => {
  return (
    <Box style={styles.container}>
      <Text size={10}>Preferences</Text>
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Manage your workspace security and how it’s members authenticate
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
  onSelection
}: PreferencesSelectionTypes) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>{label}</Text>
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          {text}
        </Text>
      </Flex>
      <Select onValueChange={onSelection} defaultValue={defaultValue}>
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
    </Flex>
  );
};
