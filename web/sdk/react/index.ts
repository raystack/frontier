import '@raystack/apsara/style.css';
import '@raystack/apsara/normalize.css';

export { ImageUpload } from './components/image-upload';
export { ViewContainer } from './components/view-container';
export { ViewHeader } from './components/view-header';
export { AuthContainer } from './components/auth-container';
export { AuthHeader } from './components/auth-header';

export { SignInView } from './views/auth/sign-in';
export { SignUpView } from './views/auth/sign-up';
export { MagicLinkView } from './views/auth/magic-link';
export { MagicLinkVerifyView } from './views/auth/magic-link-verify';
export { SubscribeView } from './views/auth/subscribe';
export { UpdatesView } from './views/auth/updates';

export { GeneralView } from './views/general';
export { PreferencesView, PreferenceRow } from './views/preferences';
export { ProfileView } from './views/profile';
export { SessionsView } from './views/sessions';
export { MembersView } from './views/members';
export { SecurityView } from './views/security';
export { ProjectsView, ProjectDetailsView } from './views/projects';
export { BillingView } from './views/billing';
export { TokensView } from './views/tokens';
export { TeamsView, TeamDetailsView } from './views/teams';
export {
  ServiceAccountsView,
  ServiceAccountDetailsView
} from './views/service-accounts';
export { PlansView } from './views/plans';
export { PatsView, PATDetailsView } from './views/pat';
export { CreateOrganizationView } from './views/create-organization';

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
