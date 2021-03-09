import { factory } from 'typeorm-seeding';
import CasbinSingleton from '../../../lib/casbin';
import { User } from '../../../model/user';

const users: any[] = [];
const createUser = async () => {
  const user = await factory(User)().create();
  users.push(user);
  return user;
};
const setupPolicies = async () => {
  const user = await createUser();
  await CasbinSingleton.enforcer.addJsonPolicy(
    { user: 'alice' },
    {
      entity: 'gojek',
      landscape: ['vn', 'id'],
      environment: 'production',
      team: 'transport'
    },
    { role: 'resource.manager' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { user: 'alice' },
    {
      entity: 'gojek'
    },
    { role: 'dwh.manager' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { user: 'frank' },
    {
      entity: 'gojek',
      landscape: 'id',
      environment: 'production',
      team: 'augur'
    },
    { role: 'resource.manager' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { user: 'bob' },
    {
      team: 'transport'
    },
    { role: 'team.admin' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { user: 'cathy' },
    {
      entity: 'gojek'
    },
    { role: 'entity.admin' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'transport' },
    {
      team: 'transport'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'augur' },
    {
      team: 'augur'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'gofinance' },
    {
      team: 'gofinance'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'transport' },
    {
      entity: 'gojek',
      privacy: 'public'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'augur' },
    {
      entity: 'gojek',
      privacy: 'public'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addJsonPolicy(
    { team: 'gofinance' },
    {
      entity: 'gofin',
      privacy: 'public'
    },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addStrPolicy(
    JSON.stringify({ team: 'de' }),
    '*',
    JSON.stringify({ role: 'super.admin' }),
    { created_by: user }
  );
};

const setupUserTeamMapping = async () => {
  const user = await createUser();

  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'alice' },
    { team: 'transport' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'bob' },
    { team: 'transport' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'dave' },
    { team: 'augur' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'frank' },
    { team: 'augur' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'ele' },
    { team: 'gofinance' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addSubjectGroupingJsonPolicy(
    { user: 'gary' },
    { team: 'de' },
    { created_by: user }
  );
};

const setupResourceProjectMapping = async () => {
  const user = await createUser();

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      resource: 'p-gojek-id-firehose-transport-123'
    },
    {
      entity: 'gojek',
      environment: 'production',
      landscape: 'id',
      team: 'transport',
      privacy: 'public'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      resource: 'p-gojek-id-firehose-augur-345'
    },
    {
      entity: 'gojek',
      environment: 'production',
      landscape: 'id',
      team: 'augur',
      privacy: 'public'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      resource: 'p-gojek-id-firehose-augur-private-345'
    },
    {
      entity: 'gojek',
      environment: 'production',
      landscape: 'id',
      team: 'augur',
      privacy: 'private'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      resource: 'p-gojek-id-beast-123'
    },
    {
      entity: 'gojek',
      environment: 'production',
      landscape: 'id',
      privacy: 'public'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      resource: 'p-gofin-id-firehose-gofinance-789'
    },
    {
      entity: 'gofin',
      environment: 'production',
      landscape: 'id',
      privacy: 'public',
      team: 'gofinance'
    },
    { created_by: user }
  );
};

const setupTeamEntityProjectMapping = async () => {
  const user = await createUser();
  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      team: 'augur'
    },
    {
      entity: 'gojek'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      team: 'transport'
    },
    {
      entity: 'gojek'
    },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addResourceGroupingJsonPolicy(
    {
      team: 'gofinance'
    },
    {
      entity: 'gofin'
    },
    { created_by: user }
  );
};

const setupPermissionRoleMapping = async () => {
  const user = await createUser();
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: '*' },
    { role: 'team.admin' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: '*' },
    { role: 'entity.admin' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: '*' },
    { role: 'super.admin' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'firehose.read' },
    { role: 'resource.viewer' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'dagger.read' },
    { role: 'resource.viewer' }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'beast.read' },
    { role: 'resource.viewer' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'firehose.write' },
    { role: 'resource.manager' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'dagger.write' },
    { role: 'resource.manager' },
    { created_by: user }
  );
  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { role: 'resource.viewer' },
    { role: 'resource.manager' },
    { created_by: user }
  );

  await CasbinSingleton.enforcer.addActionGroupingJsonPolicy(
    { action: 'beast.*' },
    { role: 'dwh.manager' },
    { created_by: user }
  );
};

export default async () => {
  await setupPolicies();
  await setupUserTeamMapping();
  await setupResourceProjectMapping();
  await setupTeamEntityProjectMapping();
  await setupPermissionRoleMapping();
  return {
    users
  };
};
