import type { AxiosInstance, AxiosResponse } from "axios";
import axios from "axios";
import type { Strategy } from "./contexts/StrategyContext";
import type { Group, ShieldClientOptions, User } from "./types";
import { Organization } from "./types/organization";

export default class Shield {
  protected readonly instance: AxiosInstance;
  private readonly options: ShieldClientOptions;
  private static classInstance?: Shield;

  static getOrCreateInstance(options: ShieldClientOptions) {
    if (!this.classInstance) {
      return new Shield(options);
    }
    return this.classInstance;
  }

  constructor(options: ShieldClientOptions) {
    this.options = options;
    this.instance = axios.create({
      baseURL: options.endpoint,
      withCredentials: true,
    });
  }

  public getAuthAtrategies = async (): Promise<
    AxiosResponse<{ strategies: Strategy[] }>
  > => {
    return await this.instance.get("/v1beta1/auth");
  };

  public getAuthStrategyEndpoint = async (
    strategy: string
  ): Promise<AxiosResponse<{ endpoint: string }>> => {
    return await this.instance.get(`/v1beta1/auth/register/${strategy}`);
  };

  public getCurrentUser = async (): Promise<AxiosResponse<{ user: User }>> => {
    return await this.instance.get("/v1beta1/users/self");
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
}
