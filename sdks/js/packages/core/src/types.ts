import React from 'react';
import { V1Beta1Feature } from '.';

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
  month: 'Monthly',
  year: 'Yearly'
} as const;

export type IntervalKeys = keyof typeof IntervalLabelMap;

export interface IntervalPricing {
  behavior: string;
  amount: Number;
  currency: string;
}

export interface IntervalPricingWithPlan extends IntervalPricing {
  planId: string;
  planName: string;
  interval: IntervalKeys;
}

export interface PlanIntervalPricing {
  slug: string;
  title: string;
  description: string;
  intervals: Record<IntervalKeys, IntervalPricingWithPlan>;
  features: Record<string, V1Beta1Feature>;
}
