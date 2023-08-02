import { V1Beta1 } from '../client/V1Beta1';
import { FrontierClient } from '../src';

// Create a class to hold the singleton instance
export default class Frontier {
  private static instance: V1Beta1 | null = null;

  private constructor() {}

  public static getInstance({ endpoint }: any): V1Beta1 {
    if (!Frontier.instance) {
      Frontier.instance = new FrontierClient({
        baseUrl: endpoint,
        baseApiParams: {
          credentials: 'include'
        }
      });
    }
    return Frontier.instance;
  }
}
