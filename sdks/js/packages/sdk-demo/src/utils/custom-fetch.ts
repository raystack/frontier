import { V1Beta1Organization } from '@raystack/frontier';
import { v4 as uuid } from 'uuid';

export const customFetch = (activeOrg?: V1Beta1Organization) => {
  return (...fetchParams: Parameters<typeof fetch>) => {
    const [url, opts] = fetchParams;
    return fetch(url, {
      ...opts,
      headers: {
        ...opts?.headers,
        'X-Request-Id': uuid(),
        'X-Frontier-Org-Id': activeOrg?.id || ''
      }
    });
  };
};
