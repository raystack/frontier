import { CachedEnforcer, newEnforcerWithClass, Util } from 'casbin';
import { JsonRoleManager } from './JsonRoleManager';

const convertJSONToStringInOrder = (json: Record<string, unknown>): string => {
  const keys = Object.keys(json).sort((a, b) => a.localeCompare(b));

  const orderedJSON = keys.reduce((acc, key) => {
    return { ...acc, [key]: json[key] };
  }, {});
  return JSON.stringify(orderedJSON);
};

type OneKey<K extends string> = Record<K, unknown>;
type JsonAttributes = Record<string, unknown>;

export class JsonEnforcer extends CachedEnforcer {
  constructor() {
    super();
    const jsonRM = new JsonRoleManager(10);
    jsonRM.addMatchingFunc(Util.keyMatchFunc);
    this.setRoleManager(jsonRM);
  }

  public async enforceJson(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    return this.enforce(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );
  }

  public async addJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    await this.addPolicy(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );

    await this.invalidateCache();
  }

  public async addSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    await this.addNamedGroupingPolicy(
      'g',
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(jsonAttributes),
      'subject'
    );

    await this.invalidateCache();
  }

  public async addResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    await this.addNamedGroupingPolicy(
      'g2',
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(jsonAttributes),
      'resource'
    );

    await this.invalidateCache();
  }

  public async addActionGroupingJsonPolicy<T extends string>(
    action: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    await this.addNamedGroupingPolicy(
      'g3',
      convertJSONToStringInOrder(action),
      convertJSONToStringInOrder(jsonAttributes),
      'action'
    );

    await this.invalidateCache();
  }
}

// newCachedEnforcer creates a cached enforcer via file or DB.
export async function newJsonEnforcer(...params: any[]): Promise<JsonEnforcer> {
  return newEnforcerWithClass(JsonEnforcer, ...params);
}
