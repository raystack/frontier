import { Api as FrontierApi } from "./frontier";

import { frontierConfig } from "../configs/frontier";

const frontierApiInstance = new FrontierApi({
  baseURL: frontierConfig.endpoint,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

export const api = frontierApiInstance.v1Beta1;
