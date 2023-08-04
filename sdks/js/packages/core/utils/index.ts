export const hasWindow = (): boolean => typeof window !== 'undefined';

export function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}
