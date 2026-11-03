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

/**
 * One reason a delete is blocked. A delete check and a refused delete both
 * give this shape. It leaves out the server's message, which is written for
 * developers.
 */
export interface DeleteBlocker {
  /** ACTIVE_SUBSCRIPTION, UNPAID_INVOICE, NEGATIVE_TOKEN_BALANCE, RESOURCES_EXIST, ... */
  type: string;
  /** The resource namespace for RESOURCES_EXIST, otherwise an id. */
  subject: string;
  /** How many exist. 0 when it comes from a refused delete. */
  count: number;
}

// What blocks a delete, for each entity.
export type DeleteBlockedPayload = {
  entity: 'organization';
  blockers: DeleteBlocker[];
};

export type DeleteBlockedContext = DeleteBlockedPayload & {
  /** What the host app calls the entity, in lower case. */
  entityLabel: string;
  close: () => void;
};

export interface FrontierProviderSlots {
  /**
   * Words the dialog that says why a delete is blocked. Return a
   * `DeleteBlockedDialog`, or nothing to show the SDK's own dialog.
   */
  deleteBlockedDialog?: Slot<DeleteBlockedContext>;
}

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
  customHeaders?: Record<string, CustomHeaderValue>;
  theme?: ThemeProviderProps;
  renderToastProvider?: boolean;
  renderThemeProvider?: boolean;
  slots?: FrontierProviderSlots;
}
