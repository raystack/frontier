import { createContext, useContext } from "react";

export type Strategy = { name: string; params: any };
export type StrategyContextProps = {
  strategies: Strategy[];
};
const initialValues: StrategyContextProps = {
  strategies: [],
};
export const StrategyContext =
  createContext<StrategyContextProps>(initialValues);
StrategyContext.displayName = "StrategyContext ";

export function useStrategyContext() {
  const context = useContext<StrategyContextProps>(StrategyContext);
  return context ? context : (initialValues as StrategyContextProps);
}
