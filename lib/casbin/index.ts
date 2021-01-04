/* eslint-disable @typescript-eslint/no-empty-function */
import TypeORMAdapter from 'typeorm-adapter';
import * as CasbinPgWatcher from 'casbin-pg-watcher';
import { join } from 'path';
import { newJsonEnforcer, JsonEnforcer } from './JsonEnforcer';

const { newWatcher } = CasbinPgWatcher;

class CasbinSingleton {
  // eslint-disable-next-line no-useless-constructor
  private constructor() {}

  public static enforcer: null | JsonEnforcer;

  public static async create(dbConnectionUrl: string) {
    if (!this.enforcer) {
      const policyAdapter = await TypeORMAdapter.newAdapter({
        type: 'postgres',
        url: dbConnectionUrl
      });

      const policyWatcher = await newWatcher({
        connectionString: dbConnectionUrl
      });
      const modelPath = join(__dirname, 'model.conf');

      this.enforcer = await newJsonEnforcer(modelPath, policyAdapter);
      this.enforcer.setWatcher(policyWatcher);
      this.enforcer.enableAutoSave(true);
      this.enforcer.enableLog(true);

      // Load the policy from DB.
      await this.enforcer.loadPolicy();
    }

    return this.enforcer;
  }
}

export default CasbinSingleton;
