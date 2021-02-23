/* eslint-disable @typescript-eslint/no-empty-function */
import TypeORMAdapter from 'typeorm-adapter';
import * as CasbinPgWatcher from 'casbin-pg-watcher';
import { join } from 'path';
import { JsonEnforcer, newJsonEnforcer } from './JsonEnforcer';
import {
  newJsonFilteredEnforcer,
  JsonFilteredEnforcer
} from './JsonFilteredEnforcer';

const { newWatcher } = CasbinPgWatcher;

class CasbinSingleton {
  // eslint-disable-next-line no-useless-constructor
  private constructor() {}

  public static enforcer: null | JsonFilteredEnforcer | JsonEnforcer = null;

  public static filtered = true;

  private static async initJsonEnforcer(
    modelPath: any,
    policyAdapter: any,
    dbConnectionUrl: string
  ) {
    const policyWatcher = await newWatcher({
      connectionString: dbConnectionUrl
    });

    this.enforcer = await newJsonEnforcer(modelPath, policyAdapter);

    this.enforcer.setWatcher(policyWatcher);
    this.enforcer.enableAutoSave(true);
    this.enforcer.enableLog(false);

    // Load the policy from DB.
    await this.enforcer.loadPolicy();
  }

  private static async initJsonFilteredEnforcer(
    modelPath: any,
    policyAdapter: any
  ) {
    this.enforcer = await newJsonFilteredEnforcer(modelPath, policyAdapter);
  }

  public static async create(dbConnectionUrl: string) {
    if (!this.enforcer) {
      const policyAdapter = await TypeORMAdapter.newAdapter({
        type: 'postgres',
        url: dbConnectionUrl
      });
      const modelPath = join(__dirname, 'model.conf');

      if (CasbinSingleton.filtered) {
        await this.initJsonFilteredEnforcer(modelPath, policyAdapter);
      } else {
        await this.initJsonEnforcer(modelPath, policyAdapter, dbConnectionUrl);
      }
    }

    return this.enforcer;
  }
}

export default CasbinSingleton;
