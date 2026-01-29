import { timestampDate, type Timestamp } from "@bufbuild/protobuf/wkt";

export function timestampToDate(timestamp?: Timestamp): Date | null {
  if (!timestamp) return null;
  return timestampDate(timestamp);
}

/**
 * Checks if a ConnectRPC Timestamp is the null time (0001-01-01T00:00:00Z)
 */
export function isNullTimestamp(timestamp?: Timestamp): boolean {
  if (!timestamp) return true;
  return Number(timestamp.seconds) <= 0;
}

export type TimeStamp = Timestamp;
