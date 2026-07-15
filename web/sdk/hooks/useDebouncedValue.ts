import { useEffect, useState } from 'react';

/**
 * Returns a debounced copy of `value` that only updates after `delay`
 * milliseconds have elapsed without `value` changing.
 *
 * Unlike a debounced *setter* (which delays writes to state), this debounces
 * the *value* itself, so it can wrap a derived/computed value (e.g. a memoized
 * request query) while the source state stays instant. The debounce runs inside
 * an effect (post-commit), which is safe under React's concurrent rendering.
 *
 * @example
 * const computedQuery = useMemo(() => transform(tableQuery), [tableQuery]);
 * const query = useDebouncedValue(computedQuery, 200);
 */
export function useDebouncedValue<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}
