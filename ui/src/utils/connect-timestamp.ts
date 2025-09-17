import { timestampDate, Timestamp } from "@bufbuild/protobuf/wkt";

/**
 * Converts a ConnectRPC Timestamp to a JavaScript Date
 */
export function timestampToDate(timestamp?: Timestamp): Date | null {
  if (!timestamp) return null;

  // Use protobuf WKT utility
  return timestampDate(timestamp);
}

/**
 * Checks if a ConnectRPC Timestamp is the null time (0001-01-01T00:00:00Z)
 */
export function isNullTimestamp(timestamp?: Timestamp): boolean {
  if (!timestamp) return true;

  // Null timestamp has very negative seconds value
  return timestamp.seconds <= 0n;
}

export type TimeStamp = Timestamp;
