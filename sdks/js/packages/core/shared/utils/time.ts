// Common timestamp interface for both protobuf libraries
interface TimestampLike {
  seconds: bigint;
  nanos?: number;
}

/**
 * Converts a protobuf Timestamp to a JavaScript Date
 * Works with both @bufbuild/protobuf and @raystack/proton timestamp types
 */
export function timestampToDate(timestamp?: TimestampLike): Date | null {
  if (!timestamp) return null;

  // Both timestamp types have the same structure: { seconds: bigint, nanos?: number }
  const seconds = Number(timestamp.seconds);
  const nanos = timestamp.nanos || 0;
  return new Date(seconds * 1000 + Math.floor(nanos / 1000000));
}

/**
 * Checks if a protobuf Timestamp is the null time (0001-01-01T00:00:00Z)
 */
export function isNullTimestamp(timestamp?: TimestampLike): boolean {
  if (!timestamp) return true;

  // Null timestamp has very negative seconds value
  return timestamp.seconds <= BigInt(0);
}

export type TimeStamp = TimestampLike;