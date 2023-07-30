import type { AxiosInstance, AxiosResponse } from 'axios';
import axios from 'axios';

import {
  FrontierClientOptions,
  Group,
  Organization,
  Strategy,
  User
} from '../shared/types';

export default class Frontier {
  protected readonly instance: AxiosInstance;
  private readonly options: FrontierClientOptions;
  private static classInstance?: Frontier;

  static getOrCreateInstance(options: FrontierClientOptions) {
    if (!this.classInstance) {
      return new Frontier(options);
    }
    return this.classInstance;
  }

  constructor(options: FrontierClientOptions) {
    this.options = options;
    this.instance = axios.create({
      baseURL: options.endpoint,
      withCredentials: true
    });
  }

  public getAuthAtrategies = async (): Promise<
    AxiosResponse<{ strategies: Strategy[] }>
  > => {
    return await this.instance.get('/v1beta1/auth');
  };

  public getAuthStrategyEndpoint = async (
    strategy: string
  ): Promise<AxiosResponse<{ endpoint: string }>> => {
    return await this.instance.get(`/v1beta1/auth/register/${strategy}`);
  };

  public getMagicLinkAuthStrategyEndpoint = async (
    email: string
  ): Promise<AxiosResponse<{ state: string }>> => {
    return await this.instance.get(
      `/v1beta1/auth/register/mailotp?email=${email}`
    );
  };

  public verifyMagicLinkAuthStrategyEndpoint = async (
    code: string,
    state: string
  ): Promise<AxiosResponse<{ state: string }>> => {
    return await this.instance.get(
      `/v1beta1/auth/callback?strategy_name=mailotp&code=${code}&state=${state}`
    );
  };

  public getCurrentUser = async (): Promise<AxiosResponse<{ user: User }>> => {
    return await this.instance.get('/v1beta1/users/self');
  };

  public getUserGroups = async (
    userId: string
  ): Promise<AxiosResponse<{ groups: Group[] }>> => {
    return await this.instance.get(`/v1beta1/users/${userId}/groups`);
  };

  public getUserOrganisations = async (
    userId: string
  ): Promise<AxiosResponse<{ organizations: Organization[] }>> => {
    return await this.instance.get(`/v1beta1/users/${userId}/organizations`);
  };

  public createOrganisation = async (
    data: any
  ): Promise<AxiosResponse<{ organization: Organization }>> => {
    return await this.instance.post(`/v1beta1/organizations`, data);
  };
}