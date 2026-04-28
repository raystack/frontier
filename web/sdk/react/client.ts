import '@raystack/apsara-v1/style.css';
import '@raystack/apsara-v1/normalize.css';

export { ImageUpload } from './components/image-upload';
export { ViewContainer } from './components/view-container';
export { ViewHeader } from './components/view-header';

export {
  SignInView,
  SignUpView,
  MagicLinkVerifyView,
  SubscribeView,
  UpdatesView
} from './views-new/auth';
export type {
  SignInViewProps,
  SignUpViewProps,
  MagicLinkVerifyViewProps,
  SubscribeViewProps,
  UpdatesViewProps
} from './views-new/auth';

export { GeneralView } from './views-new/general';
export { PreferencesView, PreferenceRow } from './views-new/preferences';
export { ProfileView } from './views-new/profile';
export { SessionsView } from './views-new/sessions';
export { MembersView } from './views-new/members';
export { SecurityView } from './views-new/security';
export { ProjectsView, ProjectDetailsView } from './views-new/projects';
export { BillingView } from './views-new/billing';
export { TokensView } from './views-new/tokens';
export { TeamsView, TeamDetailsView } from './views-new/teams';
export {
  ServiceAccountsView,
  ServiceAccountDetailsView
} from './views-new/service-accounts';
export { PlansView } from './views-new/plans';
export { PatsView, PATDetailsView } from './views-new/pat';

export { useFrontier } from './contexts/FrontierContext';
export { FrontierProvider, queryClient } from './contexts/FrontierProvider';
export { CustomizationProvider } from './contexts/CustomizationContext';

export { useTerminology } from './hooks/useTerminology';
export { useTokens } from './hooks/useTokens';
export { useBillingPermission } from './hooks/useBillingPermission';
export { useConnectQueryPolling } from './hooks/useConnectQueryPolling';
export { usePreferences } from './hooks/usePreferences';

export type {
  FrontierClientOptions,
  FrontierClientBillingOptions,
  FrontierClientCustomizationOptions
} from '../shared/types';

export { PREFERENCE_OPTIONS } from './utils/constants';

export {
  timestampToDate,
  timestampToDayjs,
  isNullTimestamp
} from '../utils/timestamp';
export type { TimeStamp } from '../utils/timestamp';
