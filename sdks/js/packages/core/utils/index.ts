export const hasWindow = (): boolean => typeof window !== 'undefined';

export function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase();
}


export const getInitials = function (name: string = '') {
  var names = name.split(' '),
    initials = names[0].substring(0, 1).toUpperCase();

  if (names.length > 1) {
    initials += names[names.length - 1].substring(0, 1).toUpperCase();
  }
  return initials;
};
