/* eslint-disable class-methods-use-this */
import * as R from 'ramda';
import { createQueryBuilder, In, Like, Raw } from 'typeorm';
import { newEnforcerWithClass } from 'casbin';
import { convertJSONToStringInOrder, JsonEnforcer } from './JsonEnforcer';
import { toLikeQuery } from '../../app/policy/util';
import IEnforcer, { JsonAttributes, OneKey, PolicyObj } from './IEnforcer';

const groupPolicyParameters = (policies: PolicyObj[]) => {
  const res = policies.reduce(
    (result: any, policy: PolicyObj) => {
      result.subjects.push(convertJSONToStringInOrder(policy.subject));
      result.resources.push(convertJSONToStringInOrder(policy.resource));
      result.actions.push(convertJSONToStringInOrder(policy.action));
      return result;
    },
    { subjects: [], resources: [], actions: [] }
  );
  return {
    subjects: <string[]>R.uniq(res.subjects),
    resources: <string[]>R.uniq(res.resources),
    actions: <string[]>R.uniq(res.actions)
  };
};

export class JsonFilteredEnforcer implements IEnforcer {
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

  private async getSubjectsForPolicySubset(subjects: string[]) {
    if (R.isEmpty(subjects)) return [];

    const subjectMappings = await createQueryBuilder()
      .select('casbin_rule.v1')
      .from('casbin_rule', 'casbin_rule')
      .where('casbin_rule.ptype = :type', { type: 'g' })
      .andWhere('casbin_rule.v0 in (:...subjects)', {
        subjects
      })
      .getRawMany();

    return R.uniq(subjectMappings.map((sM) => sM.v1).concat(subjects));
  }

  private async getResourcesForPolicySubset(resources: string[]) {
    if (R.isEmpty(resources)) return [];

    const resourceMappings = await createQueryBuilder()
      .select('casbin_rule.v1')
      .from('casbin_rule', 'casbin_rule')
      .where('casbin_rule.ptype = :type', { type: 'g2' })
      .andWhere('casbin_rule.v0 in (:...resources)', {
        resources
      })
      .getRawMany();

    const mergedResources = R.uniq(
      resourceMappings
        .map((rM) => rM.v1)
        .concat(resources)
        .map((str) => JSON.parse(str))
        .reduce((res, obj) => {
          R.keys(obj).forEach((key) => {
            res.push({ [key]: obj[key] });
          });
          return res;
        }, [])
    );

    return mergedResources.map((obj) => {
      return toLikeQuery(<Record<string, unknown>>obj);
    });
  }

  private async getActionsForPolicySubset(actions: string[]) {
    if (R.isEmpty(actions)) return [];

    const actionMappings = await createQueryBuilder()
      .select('casbin_rule.v1')
      .from('casbin_rule', 'casbin_rule')
      .where('casbin_rule.ptype = :type and casbin_rule.v0 in (:...actions)', {
        type: 'g3',
        actions
      })
      .orWhere(
        'casbin_rule.ptype = :type and casbin_rule.v0 like :allMatchPattern',
        {
          allMatchPattern: '%*%'
        }
      )
      .getRawMany();

    return R.uniq(actionMappings.map((aM) => aM.v1).concat(actions));
  }

  private async getEnforcerWithPolicies(policies: PolicyObj[]) {
    const enforcer = await this.getEnforcer();
    const { subjects, resources, actions } = groupPolicyParameters(policies);

    const [
      subjectsForPolicyFilter,
      resourcesForPolicyFilter,
      actionsForPolicyFilter
    ] = await Promise.all([
      this.getSubjectsForPolicySubset(subjects),
      this.getResourcesForPolicySubset(resources),
      this.getActionsForPolicySubset(actions)
    ]);

    const any = Like('%*%');
    const queryForPoliciesWithRegex = [{ v0: any }, { v1: any }, { v2: any }];

    const allElementsAreNonEmpty = R.all(R.complement(R.isEmpty), [
      subjectsForPolicyFilter,
      resourcesForPolicyFilter,
      actionsForPolicyFilter,
      subjects,
      resources,
      actions
    ]);
    if (allElementsAreNonEmpty) {
      await enforcer.loadFilteredPolicy({
        where: [
          {
            ptype: 'p',
            v0: In(subjectsForPolicyFilter),
            v1: Raw((alias) => `${alias} like any (array[:...resources])`, {
              resources: resourcesForPolicyFilter
            }),
            v2: In(actionsForPolicyFilter)
          },
          { ptype: 'g', v0: In(subjects) },
          { ptype: 'g2', v0: In(resources) },
          { ptype: 'g3', v0: In(actions) },
          ...queryForPoliciesWithRegex
        ]
      });
    }

    return enforcer;
  }

  public async enforceJson(
    subject: JsonAttributes,
    resource: JsonAttributes,
    action: JsonAttributes
  ): Promise<boolean> {
    const enforcer = await this.getEnforcerWithPolicies([
      { subject, resource, action }
    ]);

    const hasAccess = await enforcer.enforce(
      convertJSONToStringInOrder(subject),
      convertJSONToStringInOrder(resource),
      convertJSONToStringInOrder(action)
    );

    return hasAccess;
  }

  public async batchEnforceJson(policies: PolicyObj[]) {
    // load relevant policy subset
    const enforcer = await this.getEnforcerWithPolicies(policies);

    // enforce using Promise.all on each policy
    const enforceBatchResult = await Promise.all(
      policies.map(async (policy: PolicyObj) =>
        enforcer.enforce(
          convertJSONToStringInOrder(policy.subject),
          convertJSONToStringInOrder(policy.resource),
          convertJSONToStringInOrder(policy.action)
        )
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
): Promise<JsonFilteredEnforcer> {
  JsonFilteredEnforcer.setParams(params);
  return new JsonFilteredEnforcer();
}
