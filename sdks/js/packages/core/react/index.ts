import '@raystack/apsara/style.css';
import 'react-loading-skeleton/dist/skeleton.css';
import Amount from './components/helpers/Amount';

export { AvatarUpload } from './components/avatar-upload';
export { Container } from './components/Container';
export { Header } from './components/Header';
export { MagicLink } from './components/onboarding/magiclink';
export { MagicLinkVerify } from './components/onboarding/magiclink-verify';
export { SignIn } from './components/onboarding/signin';
export { SignUp } from './components/onboarding/signup';
export { Subscribe } from './components/onboarding/subscribe';
export { CreateOrganization } from './components/organization/create';
export { OrganizationProfile } from './components/organization/profile';
export { Window } from './components/window';

export { useFrontier } from './contexts/FrontierContext';
export { FrontierProvider } from './contexts/FrontierProvider';

export { Amount };
export { useTokens } from './hooks/useTokens';
export { useBillingPermission } from './hooks/useBillingPermission';
export { Layout } from './components/Layout';

export type {
  FrontierClientOptions,
  FrontierClientBillingOptions
} from '../shared/types';

export { PREFERENCE_OPTIONS } from './utils/constants';
