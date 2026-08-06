import { timestampDate, type Timestamp } from "@bufbuild/protobuf/wkt";
import dayjs, { type Dayjs } from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import enLocale from "dayjs/locale/en";

dayjs.extend(relativeTime);

export function timestampToDate(timestamp?: Timestamp): Date | null {
  if (!timestamp) return null;
  return timestampDate(timestamp);
}

export function timestampToDayjs(timestamp?: Timestamp): Dayjs | null {
  const date = timestampToDate(timestamp);
  return date ? dayjs(date) : null;
}

/**
 * Checks if a ConnectRPC Timestamp is the null time (0001-01-01T00:00:00Z)
 */
export function isNullTimestamp(timestamp?: Timestamp): boolean {
  if (!timestamp) return true;
  return Number(timestamp.seconds) <= 0;
}

/** Canonical admin date format (matches Apsara). */
export const DATE_FORMAT = "DD MMM YYYY";

/** Formats a Timestamp, returning "-" for missing/null-time values (which would otherwise render as 01 Jan 1970). */
export function formatTimestamp(timestamp?: Timestamp, format: string = DATE_FORMAT): string {
  if (isNullTimestamp(timestamp)) return "-";
  const date = timestampToDayjs(timestamp);
  return date ? date.format(format) : "-";
}

export type TimeStamp = Timestamp;

/*
  Invite expiry wants "5 days left"; stock "en" only says "in 5 days".
  - registered as local, so the shared "en" locale stays untouched
  - thresholds stay default: extend() installs a plugin once, so options
    passed here would lose to whichever module extends first
*/
const INVITE_LOCALE = "en-invite";

dayjs.locale(
  INVITE_LOCALE,
  {
    ...enLocale,
    relativeTime: {
      future: "%s left",
      past: "%s ago",
      s: "Less than an hour",
      m: "Less than an hour",
      mm: "Less than an hour",
      h: "1 hour",
      hh: "%d hours",
      d: "1 day",
      dd: "%d days",
      M: "1 month",
      MM: "%d months",
      y: "1 year",
      yy: "%d years",
    },
  },
  true,
);

/** Relative expiry text plus the lapsed flag. Lapsed invites show up at all because the API never filters expires_at. */
export function formatInviteExpiry(expiresAt?: Timestamp): {
  text: string;
  isExpired: boolean;
} {
  const expires = timestampToDayjs(expiresAt);
  if (!expires) return { text: "-", isExpired: false };

  return {
    text: expires.locale(INVITE_LOCALE).fromNow(),
    isExpired: !expires.isAfter(dayjs()),
  };
}
