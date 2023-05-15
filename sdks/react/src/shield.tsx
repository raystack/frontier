import type { AxiosInstance, AxiosResponse } from "axios";
import axios from "axios";
import type { Strategy } from "./contexts/StrategyContext";
import type { ShieldClientOptions, User } from "./types";

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

  public getCurrentUser = async (): Promise<AxiosResponse<{ user: User }>> => {
    return await this.instance.get("/v1beta1/users/self");
  };
}
