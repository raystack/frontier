import React from 'react';
import { BasePlan } from '../src/types';
import { ThemeProviderProps } from '@raystack/apsara';

/**
 * A render-prop slot: receives a context object (shape defined per view/slot)
 * and returns the node to render. Lets a consumer override an internal element
 * of a view while keeping the customisation on their side.
 *
 * @example
 * interface ViewSlots {
 *   addButton?: Slot<{ onClick: () => void; disabled: boolean }>;
 * }
 */
export type Slot<TContext> = (context: TContext) => React.ReactNode;

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

export interface EntityTerminologies {
  singular: string;
  plural: string;
}

export interface FrontierClientCustomizationOptions {
  terminology?: {
    organization?: EntityTerminologies;
    project?: EntityTerminologies;
    team?: EntityTerminologies;
    member?: EntityTerminologies;
    user?: EntityTerminologies;
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
  locale?: string;
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

export type CustomHeaderValue = string | (() => string);

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
  customHeaders?: Record<string, CustomHeaderValue>;
  theme?: ThemeProviderProps;
  renderToastProvider?: boolean;
  renderThemeProvider?: boolean;
}
