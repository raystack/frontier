import { useEffect, Suspense, useContext } from 'react';
import { Flex } from '@raystack/apsara';
import { useSearchParams, useNavigate } from 'react-router-dom';
import useAuthRedirect from '@/hooks/useAuthRedirect';
import { handoffAuthRejection } from '@/hooks/useAuthRejection';
import { FrontierServiceQueries, useQuery } from '@raystack/frontier/hooks';
import {
  describeAuthError,
  type AuthErrorKind
} from '@raystack/frontier/client';
import AuthContext from '@/contexts/auth';

// Where a rejection sends the user. The SDK is router agnostic, so this table
// names this application's own routes. The kind is enough to pick one:
// signup_user_exists and consent_required can only come from a signup, and
// login_user_not_found only from a login. A bad or expired flow says nothing
// about which it was, so it falls to the login page.
const REJECTION_ROUTE: Record<AuthErrorKind, string> = {
  signup_user_exists: '/signup',
  consent_required: '/signup',
  login_user_not_found: '/login',
  invalid_request: '/login',
  unknown: '/login'
};

function CallbackComponent() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setIsAuthorized } = useContext(AuthContext);

  const state = searchParams.get('state') || '';
  const code = searchParams.get('code') || '';

  const { isSuccess, isError, error } = useQuery(
    FrontierServiceQueries.authCallback,
    {
      state,
      code
    },
    // an auth rejection is final, so retrying only delays the redirect
    { enabled: !!state && !!code, retry: false }
  );

  useEffect(() => {
    if (!isSuccess) return;
    setIsAuthorized(true);
    navigate('/', { replace: true });
  }, [isSuccess, navigate, setIsAuthorized]);

  // The login gate and the consent check both reject here for oidc and mail
  // link, since neither knows the email before the redirect. This page hands
  // the rejection to the view that can act on it rather than rendering it,
  // through memory, so nothing about it reaches the URL or survives a reload.
  useEffect(() => {
    if (!isError) return;
    setIsAuthorized(false);
    const { kind, strategy } = describeAuthError(error);
    handoffAuthRejection({ kind, strategy });
    navigate(REJECTION_ROUTE[kind], { replace: true });
  }, [isError, error, navigate, setIsAuthorized]);

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      Loading...
    </Flex>
  );
}

export default function Callback() {
  useAuthRedirect();

  return (
    <Suspense>
      <CallbackComponent />
    </Suspense>
  );
}
