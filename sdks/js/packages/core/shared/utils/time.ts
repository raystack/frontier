import { Timestamp } from "@bufbuild/protobuf/wkt";

/**
 * Converts a protobuf Timestamp to a JavaScript Date
 * Works with @bufbuild/protobuf Timestamp types
 */
export function timestampToDate(timestamp?: Timestamp): Date | null {
  if (!timestamp) return null;

  // Both timestamp types have the same structure: { seconds: bigint, nanos?: number }
  const seconds = Number(timestamp.seconds);
  const nanos = timestamp.nanos || 0;
  return new Date(seconds * 1000 + Math.floor(nanos / 1000000));
}

/**
 * Checks if a protobuf Timestamp is the null time (0001-01-01T00:00:00Z)
 */
export function isNullTimestamp(timestamp?: Timestamp): boolean {
  if (!timestamp) return true;

  // Null timestamp has very negative seconds value
  return timestamp.seconds <= BigInt(0);
}

export type TimeStamp = Timestamp;