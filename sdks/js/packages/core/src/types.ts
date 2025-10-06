import React from 'react';
import { V1Beta1Feature } from '.';
import { Plan } from '@raystack/proton/frontier';

export interface Strategy {
  name: string;
  params: any;
  endpoint: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Group {
  id: string;
  name: string;
  slug: string;
  backend: string;
  resoure_type: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
}

export interface Project {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
}

export interface Role {
  id: string;
  name: string;
  types: string[];
}

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

export interface BasePlan extends Omit<Plan, 'title'> {
  features?: Record<string, string>;
  title: string;
}
