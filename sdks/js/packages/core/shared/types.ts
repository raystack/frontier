import React from 'react';
import { V1Beta1Organization } from '../api-client';
import { BasePlan } from '../src/types';
import { ThemeProviderProps } from '@raystack/apsara/v1';
export type CustomFetch = typeof fetch;

export interface FrontierClientBillingOptions {
  supportEmail?: string;
  successUrl?: string;
  cancelUrl?: string;
  hideDecimals?: boolean;
  cancelAfterTrial?: boolean;
  showPerMonthPrice?: boolean;
  tokenProductId?: string;
  basePlan?: BasePlan;
}

export interface FrontierClientCustomizationOptions {
  terminology?: {
    organization?: { singular: string; plural: string };
    project?: { singular: string; plural: string };
    team?: { singular: string; plural: string };
    member?: { singular: string; plural: string };
    user?: { singular: string; plural: string };
    appName?: string;
  };
  messages?: {
    billing?: {
      plan_change?: Record<string, string>;
    };
    general?: Record<string, string>;
  };
}

export interface FrontierClientOptions {
  endpoint: string;
  connectEndpoint?: string;
  redirectSignup?: string;
  redirectLogin?: string;
  redirectMagicLinkVerify?: string;
  callbackUrl?: string;
  dateFormat?: string;
  shortDateFormat?: string;
  billing?: FrontierClientBillingOptions;
  customization?: FrontierClientCustomizationOptions;
}

export interface InitialState {
  sessionId?: string | null;
}

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
  customFetch?: (activeOrg?: V1Beta1Organization) => CustomFetch;
  theme?: ThemeProviderProps;
}
