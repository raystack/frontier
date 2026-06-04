import dayjs, { type Dayjs } from 'dayjs';
import type { PAT } from '@raystack/proton/frontier';
import { isNullTimestamp, timestampToDayjs } from '~/utils/timestamp';

export const EXPIRY_OPTIONS = [
  { value: '1w', label: '1 week', amount: 1, unit: 'week' as const },
  { value: '1m', label: '1 month', amount: 1, unit: 'month' as const },
  { value: '3m', label: '3 months', amount: 3, unit: 'month' as const },
  { value: '6m', label: '6 months', amount: 6, unit: 'month' as const },
  { value: '12m', label: '12 months', amount: 12, unit: 'month' as const }
] as const;

export type ExpiryOption = (typeof EXPIRY_OPTIONS)[number];

/**
 * Reference timestamp for expiry math: regeneratedAt when present, else createdAt.
 */
export function getExpiryReferenceDayjs(pat?: PAT): Dayjs | null {
  if (!pat) return null;
  if (pat.regeneratedAt && !isNullTimestamp(pat.regeneratedAt)) {
    return timestampToDayjs(pat.regeneratedAt);
  }
  return timestampToDayjs(pat.createdAt);
}

export function getExpiryOptionValue(
  reference?: dayjs.Dayjs | null,
  expiresAt?: dayjs.Dayjs | null
): string {
  if (!reference || !expiresAt) return '';
  const match = EXPIRY_OPTIONS.find(option =>
    reference.add(option.amount, option.unit).isSame(expiresAt, 'day')
  );
  return match?.value ?? '';
}
