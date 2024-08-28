import { CustomFetch } from '~/shared/types';
import { V1Beta1 } from '../api-client/V1Beta1';
import { FrontierClient } from '../src';

interface FrontierOptions {
  endpoint: string;
  customFetch?: CustomFetch;
}

export const defaultFetch = (...fetchParams: Parameters<typeof fetch>) =>
  fetch(...fetchParams);

// Create a class to hold the singleton instance
export default class Frontier {
  private static instance: V1Beta1 | null = null;

  private constructor() {}

  public static getInstance({
    endpoint,
    customFetch
  }: FrontierOptions): V1Beta1 {
    if (!this.instance) {
      this.instance = new FrontierClient({
        customFetch: customFetch,
        baseUrl: endpoint,
        baseApiParams: {
          credentials: 'include'
        }
      });
    }
    return this.instance;
  }
}
