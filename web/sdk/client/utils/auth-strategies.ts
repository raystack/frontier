import type { AuthStrategy } from '@raystack/proton/frontier';
import { MAIL_OTP_STRATEGY } from './constants';

export const visibleAuthStrategies = (
  strategies: AuthStrategy[],
  excludes: string[]
): AuthStrategy[] =>
  strategies
    .filter(s => !excludes.includes(s.name))
    .sort(
      (a, b) =>
        Number(a.name === MAIL_OTP_STRATEGY) -
        Number(b.name === MAIL_OTP_STRATEGY)
    );
