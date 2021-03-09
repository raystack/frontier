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
    action: JsonAttributes,
    options: JsonAttributes
  ): void;

  addStrPolicy(
    subject: string,
    resource: string,
    action: string,
    options: JsonAttributes
  ): void;

  removeJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes,
    options: JsonAttributes
  ): void;

  addSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes,
    options: JsonAttributes
  ): void;

  removeSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes,
    options: JsonAttributes
  ): void;

  addResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes,
    options: JsonAttributes
  ): void;

  // ? Note: this will remove all policies by resource keys and then insert the new one
  upsertResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes,
    options: JsonAttributes
  ): void;

  removeAllResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    options: JsonAttributes
  ): void;

  addActionGroupingJsonPolicy<T extends string>(
    action: OneKey<T>,
    jsonAttributes: JsonAttributes,
    options: JsonAttributes
  ): void;
}

export default IEnforcer;
