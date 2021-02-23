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

  public static policyAdapter: any = null;

  public static filtered = true;

  private static async initJsonEnforcer(
    modelPath: any,
    dbConnectionUrl: string
  ) {
    const policyWatcher = await newWatcher({
      connectionString: dbConnectionUrl
    });

    this.enforcer = await newJsonEnforcer(modelPath, this.policyAdapter);

    this.enforcer.setWatcher(policyWatcher);
    this.enforcer.enableAutoSave(true);
    this.enforcer.enableLog(false);

    // Load the policy from DB.
    await this.enforcer.loadPolicy();
  }

  private static async initJsonFilteredEnforcer(modelPath: any) {
    this.enforcer = await newJsonFilteredEnforcer(
      modelPath,
      this.policyAdapter
    );
  }

  public static async create(dbConnectionUrl: string) {
    if (!this.enforcer) {
      // ? Doing this to run tests for both filtered=false/true
      if (!this.policyAdapter) {
        this.policyAdapter = await TypeORMAdapter.newAdapter({
          type: 'postgres',
          url: dbConnectionUrl
        });
      }

      const modelPath = join(__dirname, 'model.conf');
      if (CasbinSingleton.filtered) {
        await this.initJsonFilteredEnforcer(modelPath);
      } else {
        await this.initJsonEnforcer(modelPath, dbConnectionUrl);
      }
    }

    return this.enforcer;
  }
}

export default CasbinSingleton;
