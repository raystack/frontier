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

/** Canonical admin date format (matches Apsara). */
export const DATE_FORMAT = "DD MMM YYYY";

/** Formats a Timestamp, returning "-" for missing/null-time values (which would otherwise render as 01 Jan 1970). */
export function formatTimestamp(timestamp?: Timestamp, format: string = DATE_FORMAT): string {
  if (isNullTimestamp(timestamp)) return "-";
  const date = timestampToDayjs(timestamp);
  return date ? date.format(format) : "-";
}

export type TimeStamp = Timestamp;
