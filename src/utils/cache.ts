import qs from 'qs';
import Hapi from '@hapi/hapi';
import { getConnection } from 'typeorm';

export function createCacheId(
  obj: any = {},
  userId: string = '',
  name: string = ''
): string {
  let objString = qs.stringify(obj, {
    allowDots: true,
    arrayFormat: 'indices',
    encode: false
  });

  if (name) {
    objString = `${name}::${objString}`;
  }

  if (userId) {
    objString = `${userId}::${objString}`;
  }
  return objString;
}

export async function clearQueryCache(
  request: Hapi.Request,
  id: string,
  name?: string
): Promise<void> {
  // @ts-ignore
  if (request?.server?.plugins?.redis?.getKeys && id) {
    // @ts-ignore
    const getKeys = request.server.plugins.redis.getKeys;
    const pattern = name ? `${id}::${name}::*` : `${id}::*`;
    const cacheIds = getKeys ? await getKeys(pattern) : [];
    await getConnection().queryResultCache?.remove(cacheIds);
  }
}
