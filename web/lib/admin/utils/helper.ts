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

const ZERO_UUID = "00000000-0000-0000-0000-000000000000" as const;

export function isZeroUUID(uuid: string) {
  if (typeof uuid !== "string") return false;
  return uuid.toLowerCase() === ZERO_UUID;
}

export function capitalizeFirstLetter(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export function converBillingAddressToString(
  address?: { line1?: string; line2?: string; city?: string; state?: string; country?: string; postalCode?: string },
) {
  if (!address) return "";
  const { line1, line2, city, state, country, postalCode } = address;
  return [line1, line2, city, state, country, postalCode].filter(Boolean).join(", ");
}
