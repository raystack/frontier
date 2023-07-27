import { FrontierProviderProps } from '../../shared/types';
import { FrontierContextProvider } from './FrontierContext';
import { withMaxAllowedInstancesGuard } from './useMaxAllowedInstancesGuard';

export const multipleFrontierProvidersError =
  "Frontier: You've added multiple <FrontierProvider> components in your React component tree. Wrap your components in a single <FrontierProvider>.";

const FrontierProviderBase = (props: FrontierProviderProps) => {
  const { children, initialState, config, ...options } = props;
  return (
    <FrontierContextProvider
      initialState={initialState}
      config={config}
      {...options}
    >
      {children}
    </FrontierContextProvider>
  );
};

export const FrontierProvider =
  withMaxAllowedInstancesGuard<FrontierProviderProps>(
    FrontierProviderBase,
    'FrontierProvider',
    multipleFrontierProvidersError
  );
FrontierProvider.displayName = 'FrontierProvider';
