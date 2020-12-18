import { CachedEnforcer, newEnforcerWithClass, Util } from 'casbin';
import { JsonRoleManager } from './JsonRoleManager';

const convertJSONToStringInOrder = (json: Record<string, unknown>): string => {
  const keys = Object.keys(json).sort((a, b) => a.localeCompare(b));

  const orderedJSON = keys.reduce((acc, key) => {
    return { ...acc, key: json[key] };
  }, {});
  return JSON.stringify(orderedJSON);
};

type OneKey<K extends string> = Record<K, unknown>;
type JsonAttributes = Record<string, unknown>;

export class JsonEnforcer extends CachedEnforcer {
  constructor() {
    super();
    const jsonRM = new JsonRoleManager(10);
    jsonRM.addMatchingFunc(Util.keyMatch2Func);
    this.setRoleManager(jsonRM);
  }

  public enforceJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    this.enforce(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );
  }

  public addJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    this.addPolicy(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );

    this.invalidateCache();
  }

  public addSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    this.addNamedGroupingPolicy(
      'g',
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(jsonAttributes),
      'subject'
    );

    this.invalidateCache();
  }

  public addResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    this.addNamedGroupingPolicy(
      'g2',
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(jsonAttributes),
      'resource'
    );

    this.invalidateCache();
  }

  public addActionGroupingJsonPolicy<T extends string>(
    action: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    this.addNamedGroupingPolicy(
      'g3',
      convertJSONToStringInOrder(action),
      convertJSONToStringInOrder(jsonAttributes),
      'action'
    );

    this.invalidateCache();
  }
}

// newCachedEnforcer creates a cached enforcer via file or DB.
export async function newJsonEnforcer(...params: any[]): Promise<JsonEnforcer> {
  return newEnforcerWithClass(JsonEnforcer, ...params);
}
