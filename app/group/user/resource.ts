const create = async (
  groupIdentifier: string,
  userIdentifier: string,
  policy: any
) => {
  // create
  console.log('create => ', groupIdentifier, userIdentifier, policy);
};

const update = async (
  groupIdentifier: string,
  userIdentifier: string,
  policy: any
) => {
  // update
  console.log('update => ', groupIdentifier, userIdentifier, policy);
};

const remove = async (
  groupIdentifier: string,
  userIdentifier: string,
  policy: any
) => {
  // remove
  console.log('remove => ', groupIdentifier, userIdentifier, policy);
};

export const operation = async (
  groupIdentifier: string,
  userIdentifier: string,
  payload: any
) => {
  const { policies = [] } = payload;
  console.log('operation => ', groupIdentifier, userIdentifier, policies);
  await Promise.all(
    policies.map((policy: any) => {
      if (policy.operation === 'create') {
        return create(groupIdentifier, userIdentifier, policy);
      }
      if (policy.operation === 'update') {
        return update(groupIdentifier, userIdentifier, policy);
      }
      if (policy.operation === 'delete') {
        return remove(groupIdentifier, userIdentifier, policy);
      }
      return Promise.resolve();
    })
  );
};
