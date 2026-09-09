import { useEffect, useState } from 'react';
import type { AuthRejection } from '@raystack/frontier/client';

let pending: AuthRejection | undefined;

export const handoffAuthRejection = (rejection: AuthRejection) => {
  pending = rejection;
};

export const useAuthRejection = (): AuthRejection | undefined => {
  const [rejection] = useState(() => pending);
  useEffect(() => {
    pending = undefined;
  }, []);
  return rejection;
};
