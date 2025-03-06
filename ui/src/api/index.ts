import { Api as FrontierApi } from "./frontier";

const frontierEndpoint =
  process.env.NEXT_PUBLIC_FRONTIER_URL || "/frontier-api";

const frontierApiInstance = new FrontierApi({
  baseURL: frontierEndpoint,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

export const api = frontierApiInstance.v1Beta1;
