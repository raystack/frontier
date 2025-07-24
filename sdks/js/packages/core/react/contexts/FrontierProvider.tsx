import { ThemeProvider } from '@raystack/apsara/v1';
import { FrontierProviderProps } from '../../shared/types';
import { FrontierContextProvider } from './FrontierContext';
import { withMaxAllowedInstancesGuard } from './useMaxAllowedInstancesGuard';
import { I18nProvider } from '../components/I18nProvider';

export const multipleFrontierProvidersError =
  "Frontier: You've added multiple <FrontierProvider> components in your React component tree. Wrap your components in a single <FrontierProvider>.";

export const FrontierProvider = (props: FrontierProviderProps) => {
  const { children, initialState, config, theme, translations, ...options } = props;
  return (
    <FrontierContextProvider
      initialState={initialState}
      config={config}
      {...options}
    >
      <I18nProvider resources={translations}>
        <ThemeProvider {...theme}>{children}</ThemeProvider>
      </I18nProvider>
    </FrontierContextProvider>
  );
};
FrontierProvider.displayName = 'FrontierProvider';

export const FrontierProviderGaurd =
  withMaxAllowedInstancesGuard<FrontierProviderProps>(
    FrontierProvider,
    'FrontierProvider',
    multipleFrontierProvidersError
  );
