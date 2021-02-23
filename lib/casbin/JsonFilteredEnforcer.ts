/* eslint-disable class-methods-use-this */
import { createQueryBuilder, In, Like } from 'typeorm';
import { newEnforcerWithClass } from 'casbin';
import {
  convertJSONToStringInOrder,
  JsonEnforcer,
  JsonAttributes,
  OneKey,
  PolicyObj
} from './JsonEnforcer';

export class JsonFilteredEnforcer extends JsonEnforcer {
  public static params: any[];

  public static setParams(params: any[]) {
    this.params = params;
  }

  private async getEnforcer() {
    const enforcer = await newEnforcerWithClass(
      JsonEnforcer,
      ...JsonFilteredEnforcer.params
    );
    enforcer.enableAutoSave(true);
    enforcer.enableLog(false);
    return enforcer;
  }

  private async loadPolicySubset(policyObj: any, enforcer: JsonEnforcer) {
    const rawSubjects = await createQueryBuilder()
      .select('casbin_rule.v1')
      .from('casbin_rule', 'casbin_rule')
      .where('casbin_rule.ptype = :type', { type: 'g' })
      .andWhere('casbin_rule.v0 like :subject', {
        subject: convertJSONToStringInOrder(policyObj.subject)
      })
      .getRawMany();
    const subjects = rawSubjects
      .map((rG) => rG.v1)
      .concat(convertJSONToStringInOrder(policyObj.subject));

    const anyAction = Like('%*%');

    await enforcer.loadFilteredPolicy({
      where: [
        { ptype: 'p', v0: In(subjects) },
        { ptype: 'g', v0: convertJSONToStringInOrder(policyObj.subject) },
        { ptype: 'g2', v0: convertJSONToStringInOrder(policyObj.resource) },
        { ptype: 'g3', v0: convertJSONToStringInOrder(policyObj.action) },
        { ptype: 'g3', v0: anyAction }
      ]
    });
  }

  public async enforceJson(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ): Promise<boolean> {
    // intantiate new enforcer
    const enforcer = await this.getEnforcer();

    // load filtered policy
    await this.loadPolicySubset({ subject, resource, action }, enforcer);
    // enforceJson

    const hasAccess = await enforcer.enforce(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );

    return hasAccess;
  }

  public async batchEnforceJson(policies: PolicyObj[]) {
    const enforceBatchResult = await Promise.all(
      policies.map(async (policy: PolicyObj) =>
        this.enforceJson(policy.subject, policy.resource, policy.action)
      )
    );
    return enforceBatchResult;
  }

  public async addJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.addPolicy(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );
  }

  public async addStrPolicy(subject: string, resource: string, action: string) {
    const enforcer = await this.getEnforcer();
    await enforcer.addPolicy(subject, resource, action);
  }

  public async removeJsonPolicy(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.removeFilteredNamedPolicy(
      'p',
      0,
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );
  }

  public async addSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.addNamedGroupingPolicy(
      'g',
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(jsonAttributes),
      'subject'
    );
  }

  public async removeSubjectGroupingJsonPolicy<T extends string>(
    subject: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.removeFilteredNamedGroupingPolicy(
      'g',
      0,
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(jsonAttributes),
      'subject'
    );
  }

  public async addResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.addNamedGroupingPolicy(
      'g2',
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(jsonAttributes),
      'resource'
    );
  }

  // ? Note: this will remove all policies by resource keys and then insert the new one
  public async upsertResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    await this.removeAllResourceGroupingJsonPolicy(resource);
    await this.addResourceGroupingJsonPolicy(resource, jsonAttributes);
  }

  public async removeAllResourceGroupingJsonPolicy<T extends string>(
    resource: OneKey<T>
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.removeFilteredNamedGroupingPolicy(
      'g2',
      0,
      convertJSONToStringInOrder(resource)
    );
  }

  public async addActionGroupingJsonPolicy<T extends string>(
    action: OneKey<T>,
    jsonAttributes: JsonAttributes
  ) {
    const enforcer = await this.getEnforcer();
    await enforcer.addNamedGroupingPolicy(
      'g3',
      convertJSONToStringInOrder(action),
      convertJSONToStringInOrder(jsonAttributes),
      'action'
    );
  }
}

export async function newJsonFilteredEnforcer(
  ...params: any[]
): Promise<JsonEnforcer> {
  JsonFilteredEnforcer.setParams(params);
  return newEnforcerWithClass(JsonFilteredEnforcer, ...params);
}
