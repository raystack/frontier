import TypeORMAdapter from 'typeorm-adapter';
import path from 'path';
import { newJsonEnforcer, JsonEnforcer } from './casbin/JsonEnforcer';

type User = Record<'user', string>;
type Team = Record<'team', string>;
type Entity = Record<'entity', string>;
type Action = Record<'action', string>;
type Resource = Record<'resource', string>;
type Role = Record<'role', string>;
type ResourceAttribtues = {
  entity?: string;
  environment?: 'production' | 'integration';
  landscape?: string;
  team?: string;
  privacy?: 'public' | 'private';
};

class Iam {
  private jsonEnforcer;

  constructor(jsonEnforcer: JsonEnforcer) {
    this.jsonEnforcer = jsonEnforcer;
  }

  public async hasAccess(
    subject: User,
    resource: ResourceAttribtues,
    action: Action
  ): Promise<boolean> {
    return this.jsonEnforcer.enforceJsonPolicy(subject, resource, action);
  }

  public async mapUserToTeam(user: User, team: Team) {
    return this.jsonEnforcer.addSubjectGroupingJsonPolicy(user, team);
  }

  public async mapActionToRole(action: Action, role: Role) {
    return this.jsonEnforcer.addActionGroupingJsonPolicy(action, role);
  }

  public async mapResourceToAttributes(
    resource: Resource,
    attributes: ResourceAttribtues
  ) {
    return this.jsonEnforcer.addResourceGroupingJsonPolicy(
      resource,
      attributes
    );
  }

  public async mapTeamToEntity(team: Team, entity: Entity) {
    return this.jsonEnforcer.addResourceGroupingJsonPolicy(team, entity);
  }
}

const setupEnforcer = async (dbConnectionUrl: string) => {
  const policyAdapter = await TypeORMAdapter.newAdapter({
    type: 'postgres',
    url: dbConnectionUrl
  });
  const modelPath = path.resolve('./casbin/model.conf');
  const jsonEnforcer = await newJsonEnforcer(modelPath, policyAdapter);
  jsonEnforcer.enableAutoSave(true);
  jsonEnforcer.enableLog(true);

  // Load the policy from DB.
  await jsonEnforcer.loadPolicy();
  return jsonEnforcer;
};

const createIamInstance = async (dbConnectionUrl: string) => {
  const jsonEnforcer = await setupEnforcer(dbConnectionUrl);
  return new Iam(jsonEnforcer);
};

export default createIamInstance;
