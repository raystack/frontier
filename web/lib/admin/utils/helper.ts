export function reduceByKey<T extends Record<string, unknown>>(
  data: T[],
  key: keyof T
): Record<string, T> {
  return data.reduce(
    (acc, value) => ({
      ...acc,
      [String(value[key])]: value,
    }),
    {} as Record<string, T>
  );
}
