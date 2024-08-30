import React from 'react';
import { V1Beta1Feature, V1Beta1Plan } from '.';

export interface FrontierClientOptions {
  endpoint?: string;
  redirectSignup?: string;
  redirectLogin?: string;
}

export interface InitialState {
  sessionId?: string | null;
}

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
}

export const IntervalLabelMap = {
  daily: 'Daily',
  month: 'Monthly',
  year: 'Yearly'
} as const;

export type IntervalKeys = keyof typeof IntervalLabelMap;

export interface IntervalPricing {
  amount: number;
  currency: string;
}

export interface PlanMetadata extends Record<string, any> {
  weightage?: number;
}

export interface IntervalPricingWithPlan extends IntervalPricing {
  planId: string;
  planName: string;
  interval: IntervalKeys;
  weightage: number;
  features: Record<string, V1Beta1Feature>;
  trial_days: string;
  productNames: string[];
}

export interface PlanIntervalPricing {
  slug: string;
  title: string;
  description: string;
  intervals: Record<IntervalKeys, IntervalPricingWithPlan>;
  weightage: number;
}

export interface PaymentMethodMetadata extends Record<string, any> {
  default?: boolean;
}

export interface BasePlan extends Omit<V1Beta1Plan, 'title'> {
  features?: Record<string, string>;
  title: string;
}
