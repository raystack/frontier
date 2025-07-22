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

export interface FrontierClientOptions {
  endpoint: string;
  redirectSignup?: string;
  redirectLogin?: string;
  redirectMagicLinkVerify?: string;
  callbackUrl?: string;
  dateFormat?: string;
  shortDateFormat?: string;
  billing?: FrontierClientBillingOptions;
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
