export type OneKey<K extends string> = Record<K, unknown>;
export type JsonAttributes = Record<string, unknown>;

export type PolicyObj = {
  subject: JsonAttributes;
  resource: JsonAttributes;
  action: JsonAttributes;
};

interface IEnforcer {
  enforceJson(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ): Promise<boolean>;

  batchEnforceJson(policies: PolicyObj[]): Promise<boolean[]>;

  addJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ): void;

  addStrPolicy(subject: string, resource: string, action: string): void;

  removeJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ): void;

  addSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ): void;

  removeSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ): void;

  addResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ): void;

  // ? Note: this will remove all policies by resource keys and then insert the new one
  upsertResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ): void;

  removeAllResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>
  ): void;

  addActionGroupingJsonPolicy<T extends string>(
    action: OneKey<T>,
    jsonAttributes: JsonAttributes
  ): void;
}

export default IEnforcer;
