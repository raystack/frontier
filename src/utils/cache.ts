import qs from 'qs';
export function createCacheId(
  obj: any = {},
  userId: string = '',
  prefix: string = ''
): string {
  let objString = qs.stringify(obj, {
    allowDots: true,
    arrayFormat: 'indices',
    encode: false
  });
  if (userId) {
    objString = `${userId}::${objString}`;
  }

  if (prefix) {
    objString = `${prefix}::${objString}`;
  }
  return objString;
}
