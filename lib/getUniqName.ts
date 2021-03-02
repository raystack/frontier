import _ from 'lodash';

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

export default getUniqName;
