import type { InitialState } from "./base";

export type ShieldProviderProps = {
  endpoint: string;
  children: React.ReactNode;
  initialState?: InitialState;
};
