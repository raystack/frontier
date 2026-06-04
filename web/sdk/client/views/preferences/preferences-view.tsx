import { Flex, Select } from '@raystack/apsara';
import { SunIcon, MoonIcon, GearIcon } from '@radix-ui/react-icons';
import { BellIcon, BellSlashIcon } from '@raystack/apsara/icons';
import { ViewContainer } from '~/client/components/view-container';
import { ViewHeader } from '~/client/components/view-header';
import { usePreferences } from '~/client/hooks/usePreferences';
import { PREFERENCE_OPTIONS } from '~/client/utils/constants';
import { PreferenceRow } from './components/preference-row';
import { useTheme } from '@raystack/apsara';
import styles from './preferences-view.module.css';
import { ReactNode } from 'react';

const THEME_OPTIONS = {
  light: {
    label: 'Light',
    icon: <SunIcon />
  },
  dark: {
    label: 'Dark',
    icon: <MoonIcon />
  },
  system: {
    label: 'System',
    icon: <GearIcon />
  }
}
type Theme = keyof typeof THEME_OPTIONS;

interface PreferencesViewProps {
  children?: ReactNode;
  /**
   * The theme to use for Theme Select.
   * If not provided, the theme will be fetched from ThemeProvider.
   */
  theme?: Theme;
  /**
   * The callback to call when the theme is changed.
   * If not provided, the theme will be set in the ThemeProvider.
   */
  onThemeChange?: (theme: Theme) => void;
}
export function PreferencesView({ children, theme: providedTheme, onThemeChange }: PreferencesViewProps) {
  const { theme, setTheme } = useTheme();
  const { preferences, isLoading, isFetching, updatePreferences } =
    usePreferences({});
  const computedTheme = providedTheme ?? theme;

  const handleThemeChange = (theme: string) => {
    if (onThemeChange) {
      onThemeChange(theme as Theme);
    } else {
      setTheme(theme);
    }
  }

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
          <Select defaultValue={computedTheme} onValueChange={handleThemeChange}>
            <Select.Trigger className={styles.selectTrigger}>
              <Select.Value placeholder="Theme" />
            </Select.Trigger>
            <Select.Content>
              {Object.entries(THEME_OPTIONS).map(([value, { label, icon }]) => (
                <Select.Item key={value} value={value} leadingIcon={icon}>
                  {label}
                </Select.Item>
              ))}
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
