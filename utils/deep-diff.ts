const { diff } = require('deep-diff');

export const delta = (
  previous = {},
  current = {},
  options?: { exclude: string[] }
) => {
  return diff(previous, current).filter((i: any) =>
    (options?.exclude || []).every((x: any) => i.path.indexOf(x) === -1)
  );
};
