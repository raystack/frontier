import { ThemeProvider } from '@raystack/apsara';
import { FrontierProviderProps } from '../../shared/types';
import { FrontierContextProvider } from './FrontierContext';
import { CustomizationProvider } from './CustomizationContext';
import { withMaxAllowedInstancesGuard } from './useMaxAllowedInstancesGuard';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ComponentType, ReactNode, useMemo } from 'react';
import { createFetchWithCreds } from '../utils/fetch';
import { Toast } from '@raystack/apsara-v1';

export const multipleFrontierProvidersError =
  "Frontier: You've added multiple <FrontierProvider> components in your React component tree. Wrap your components in a single <FrontierProvider>.";

export const queryClient = new QueryClient();

export const FrontierProvider = (props: FrontierProviderProps) => {
  const { children, initialState, config, theme, customHeaders, renderThemeProvider = true, renderToastProvider = true, ...options } =
    props;

  const transport = useMemo(
    () =>
      createConnectTransport({
        baseUrl: config.connectEndpoint || '/frontier-connect',
        fetch: createFetchWithCreds(customHeaders)
      }),
    [config.connectEndpoint, customHeaders]
  );

  return (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>
        <CustomizationProvider config={config.customization}>
          <FrontierContextProvider
            initialState={initialState}
            config={config}
            {...options}
          >
            <OptionalProvider provider={ThemeProvider} shouldRender={renderThemeProvider} providerProps={theme}>
              <OptionalProvider provider={Toast.Provider} shouldRender={renderToastProvider}>
                {children}
              </OptionalProvider>
            </OptionalProvider>
          </FrontierContextProvider>
        </CustomizationProvider>
      </TransportProvider>
    </QueryClientProvider>
  );
};
FrontierProvider.displayName = 'FrontierProvider';

export const FrontierProviderGaurd =
  withMaxAllowedInstancesGuard<FrontierProviderProps>(
    FrontierProvider,
    'FrontierProvider',
    multipleFrontierProvidersError
  );

export const OptionalProvider = <T extends { children?: ReactNode }>({
  children,
  provider: Provider,
  shouldRender = true,
  providerProps
}: {
  children?: ReactNode;
  provider: ComponentType<T>;
  shouldRender?: boolean;
  providerProps?: Omit<T, 'children'>;
}) => {
  if (shouldRender) return <Provider {...(providerProps as T)}>{children}</Provider>;
  return children
};