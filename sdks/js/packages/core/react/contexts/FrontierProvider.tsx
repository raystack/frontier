import { ThemeProvider } from '@raystack/apsara/v1';
import { FrontierProviderProps } from '../../shared/types';
import { FrontierContextProvider } from './FrontierContext';
import { withMaxAllowedInstancesGuard } from './useMaxAllowedInstancesGuard';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { useMemo } from 'react';

export const multipleFrontierProvidersError =
  "Frontier: You've added multiple <FrontierProvider> components in your React component tree. Wrap your components in a single <FrontierProvider>.";

const queryClient = new QueryClient();

export const FrontierProvider = (props: FrontierProviderProps) => {
  const { children, initialState, config, theme, ...options } = props;
  
  const transport = useMemo(() => createConnectTransport({
    baseUrl: config.connectEndpoint || '/frontier-connect',
  }), [config.connectEndpoint]);
  
  return (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>
        <FrontierContextProvider
          initialState={initialState}
          config={config}
          {...options}
        >
          <ThemeProvider {...theme}>{children}</ThemeProvider>
        </FrontierContextProvider>
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
