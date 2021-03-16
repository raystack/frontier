const { diff } = require('deep-diff');

export const delta = (
  previous = {},
  current = {},
  options?: { exclude: string[] }
) => {
  return diff(previous, current).filter((i: any) => {
    return (options?.exclude || []).every(
      (x: any) => i.path && i.path.indexOf(x) === -1
    );
  });
};
