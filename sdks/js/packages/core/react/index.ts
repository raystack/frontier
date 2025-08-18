import '@raystack/apsara/style.css';

export { AvatarUpload } from './components/avatar-upload';
export { Container } from './components/Container';
export { Header } from './components/Header';
export { MagicLink } from './components/onboarding/magiclink';
export { MagicLinkVerify } from './components/onboarding/magiclink-verify';
export { SignIn } from './components/onboarding/signin';
export { SignUp } from './components/onboarding/signup';
export { Updates } from './components/onboarding/updates';
export { Subscribe } from './components/onboarding/subscribe';
export { CreateOrganization } from './components/organization/create';
export { OrganizationProfile } from './components/organization/profile';
export { Window } from './components/window';

export { useFrontier } from './contexts/FrontierContext';
export { FrontierProvider } from './contexts/FrontierProvider';
export { 
  CustomizationProvider
} from './contexts/CustomizationContext';


export { useTokens } from './hooks/useTokens';
export { useBillingPermission } from './hooks/useBillingPermission';
export { usePreferences } from './hooks/usePreferences';
export { Layout } from './components/Layout';

export type {
  FrontierClientOptions,
  FrontierClientBillingOptions,
  FrontierClientCustomizationOptions
} from '../shared/types';

export { PREFERENCE_OPTIONS } from './utils/constants';
