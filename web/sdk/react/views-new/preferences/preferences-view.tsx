import { Flex, Select } from '@raystack/apsara-v1';
import { SunIcon, MoonIcon, GearIcon } from '@radix-ui/react-icons';
import { BellIcon, BellSlashIcon } from '@raystack/apsara-v1/icons';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import { usePreferences } from '~/react/hooks/usePreferences';
import { PREFERENCE_OPTIONS } from '~/react/utils/constants';
import { PreferenceRow } from './components/preference-row';
import { useTheme } from '@raystack/apsara';
import styles from './preferences-view.module.css';
import { ReactNode } from 'react';

interface PreferencesViewProps {
  children?: ReactNode;
}
export function PreferencesView({ children }: PreferencesViewProps) {
  const { theme, setTheme } = useTheme();
  const { preferences, isLoading, isFetching, updatePreferences } =
    usePreferences({});

  const newsletterValue =
    preferences?.[PREFERENCE_OPTIONS.NEWSLETTER]?.value ?? 'false';

  return (
    <ViewContainer>
      <ViewHeader
        title="Preferences"
        description="Manage members for this domain."
      />
      <Flex direction="column">
        <PreferenceRow
          title="Theme"
          description="Customise your interface color scheme."
        >
          <Select defaultValue={theme} onValueChange={setTheme}>
            <Select.Trigger className={styles.selectTrigger}>
              <Select.Value placeholder="Theme" />
            </Select.Trigger>
            <Select.Content>
              <Select.Item value="light" leadingIcon={<SunIcon />}>
                Light
              </Select.Item>
              <Select.Item value="dark" leadingIcon={<MoonIcon />}>
                Dark
              </Select.Item>
              <Select.Item value="system" leadingIcon={<GearIcon />}>
                System
              </Select.Item>
            </Select.Content>
          </Select>
        </PreferenceRow>

        <PreferenceRow
          title="Updates, News & Events"
          description="Stay informed on new features, improvements, and key updates."
          isLoading={isFetching}
        >
          <Select
            defaultValue={newsletterValue}
            onValueChange={value => {
              updatePreferences([
                { name: PREFERENCE_OPTIONS.NEWSLETTER, value }
              ]);
            }}
            disabled={isLoading}
          >
            <Select.Trigger className={styles.selectTrigger}>
              <Select.Value placeholder="Newsletter" />
            </Select.Trigger>
            <Select.Content>
              <Select.Item value="true" leadingIcon={<BellIcon />}>
                Subscribed
              </Select.Item>
              <Select.Item value="false" leadingIcon={<BellSlashIcon />}>
                Unsubscribed
              </Select.Item>
            </Select.Content>
          </Select>
        </PreferenceRow>
        {children}
      </Flex>
    </ViewContainer>
  );
}
