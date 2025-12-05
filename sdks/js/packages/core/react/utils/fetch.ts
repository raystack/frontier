import { CustomHeaderValue } from '../../shared/types';

export const createFetchWithCreds = (
  customHeaders?: Record<string, CustomHeaderValue>
): typeof fetch => {
  return (input, init) => {
    const headers = new Headers(init?.headers);
    if (customHeaders) {
      for (const [key, value] of Object.entries(customHeaders)) {
        const headerValue = typeof value === 'function' ? value() : value;
        headers.set(key, headerValue);
      }
    }
    return fetch(input, {
      ...init,
      credentials: 'include',
      headers
    });
  };
};
