import { timestampDate, type Timestamp } from "@bufbuild/protobuf/wkt";
import dayjs, { type Dayjs } from "dayjs";

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

/** Canonical date format used across the admin UI (matches Apsara). */
export const DATE_FORMAT = "DD MMM YYYY";

/**
 * Formats a Timestamp for display, returning "-" for missing or null-time
 * values (e.g. { seconds: 0 }, which would otherwise render as 01 Jan 1970).
 */
export function formatTimestamp(timestamp?: Timestamp, format: string = DATE_FORMAT): string {
  if (isNullTimestamp(timestamp)) return "-";
  const date = timestampToDayjs(timestamp);
  return date ? date.format(format) : "-";
}

/** Table cell renderer that formats a Timestamp column value via formatTimestamp. */
export function timestampCell({ getValue, }: { getValue: () => unknown; }): string {
  return formatTimestamp(getValue() as Timestamp | undefined);
}

export type TimeStamp = Timestamp;
