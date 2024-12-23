import { ThemeProvider } from '@raystack/apsara/v1';
import { FrontierProviderProps } from '../../shared/types';
import { FrontierContextProvider } from './FrontierContext';
import { withMaxAllowedInstancesGuard } from './useMaxAllowedInstancesGuard';

export const multipleFrontierProvidersError =
  "Frontier: You've added multiple <FrontierProvider> components in your React component tree. Wrap your components in a single <FrontierProvider>.";

export const FrontierProvider = (props: FrontierProviderProps) => {
  const { children, initialState, config, theme, ...options } = props;
  return (
    <FrontierContextProvider
      initialState={initialState}
      config={config}
      {...options}
    >
      <ThemeProvider {...theme}>{children}</ThemeProvider>
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
