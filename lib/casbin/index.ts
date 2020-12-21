import TypeORMAdapter from 'typeorm-adapter';
import path from 'path';
import { newJsonEnforcer, JsonEnforcer } from './JsonEnforcer';

export const enforcerContainer: Record<'enforcer', null | JsonEnforcer> = {
  enforcer: null
};

export default async (dbConnectionUrl: string) => {
  if (!enforcerContainer.enforcer) {
    const policyAdapter = await TypeORMAdapter.newAdapter({
      type: 'postgres',
      url: dbConnectionUrl
    });
    // ! handle path properly
    const modelPath = path.resolve('./lib/casbin/model.conf');
    enforcerContainer.enforcer = await newJsonEnforcer(
      modelPath,
      policyAdapter
    );
    enforcerContainer.enforcer.enableAutoSave(true);
    enforcerContainer.enforcer.enableLog(true);

    // Load the policy from DB.
    await enforcerContainer.enforcer.loadPolicy();
  }
};
