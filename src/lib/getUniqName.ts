import _ from 'lodash';
import * as R from 'ramda';

const getUniqName = async (
  basename: string,
  key: string,
  Model: any
): Promise<string> => {
  const strippedName = basename.replace(/\s/g, '_').toLowerCase();
  const alreadyExists = await Model.findOne({ [key]: strippedName });
  if (!alreadyExists) return strippedName;

  const usernameWithNumberId = _.uniqueId(`${strippedName}_`);
  return getUniqName(usernameWithNumberId, key, Model);
};

export const validateUniqName = (name: string) => {
  const hasWhiteSpace = /\s/g.test(name);
  if (hasWhiteSpace) throw new Error('white spaces are not allowed');

  if (R.isEmpty(name) || R.isNil(name)) throw new Error(`can't be empty`);

  return true;
};

export default getUniqName;
