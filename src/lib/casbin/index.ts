/* eslint-disable @typescript-eslint/no-empty-function */
import TypeORMAdapter from 'typeorm-adapter';
import { join } from 'path';
import {
  newJsonFilteredEnforcer,
  JsonFilteredEnforcer
} from './JsonFilteredEnforcer';
import ormConfig from '../../config/ormconfig';

class CasbinSingleton {
  // eslint-disable-next-line no-useless-constructor
  private constructor() {}

  public static enforcer: null | JsonFilteredEnforcer = null;

  public static policyAdapter: any = null;

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
          type: ormConfig.type,
          url: ormConfig.url,
          subscribers: ormConfig.subscribers
        });
      }

      const modelPath = join(__dirname, 'model.conf');
      await this.initJsonFilteredEnforcer(modelPath);
    }

    return this.enforcer;
  }
}

export default CasbinSingleton;
