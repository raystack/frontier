import { useEffect, useState } from 'react';
import type { AuthErrorKind } from '@raystack/frontier/client';

export type AuthRejection = {
  kind: AuthErrorKind;
  strategy?: string;
};

// The callback page hands a rejection to the login or signup page through
// memory, not the URL or the history entry. It survives the in-app navigation
// between the two and nothing else: a reload, a new tab or a shared link start
// clean, and no error text or code is ever visible in the address bar.
let pending: AuthRejection | undefined;

export const handoffAuthRejection = (rejection: AuthRejection) => {
  pending = rejection;
};

// Read once on mount, cleared once mounted. The read is idempotent because
// StrictMode runs state initializers twice; the clear is in an effect so a
// later re-render does not lose the value.
export const useAuthRejection = (): AuthRejection | undefined => {
  const [rejection] = useState(() => pending);
  useEffect(() => {
    pending = undefined;
  }, []);
  return rejection;
};
