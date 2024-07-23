import React from 'react';
import { BasePlan } from '~/src/types';

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
  theme?: 'dark' | 'light';
  endpoint?: string;
  redirectSignup?: string;
  redirectLogin?: string;
  redirectMagicLinkVerify?: string;
  callbackUrl?: string;
  dateFormat?: string;
  shortDateFormat?: string;
  billing?: FrontierClientBillingOptions;
  messages?: {
    billing?: {
      plan_change?: Record<string, string>;
    };
  };
}

export interface InitialState {
  sessionId?: string | null;
}

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
}
