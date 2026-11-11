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
  /**
   * Check what blocks an organization, project or account delete when the
   * delete button is clicked, and show the blockers in a dialog. The server
   * must serve CheckOrganizationDelete, CheckProjectDelete and
   * CheckCurrentUserDelete. Off by default: the delete button opens the
   * confirm dialog directly, and the organization page checks on load and
   * disables the button while blocked.
   */
  enableDeleteCheck?: boolean;
}

export interface InitialState {
  sessionId?: string | null;
}

export type CustomHeaderValue = string | (() => string);

export interface DeleteBlocker {
  type: string;
  subject: string;
  count?: number;
}

export type DeleteBlockedPayload =
  | { entity: 'organization' | 'project'; blockers: DeleteBlocker[] }
  | {
      entity: 'user';
      organizationsToDelete: {
        id: string;
        title: string;
        blockers: DeleteBlocker[];
      }[];
      soleOwnerOf: { id: string; title: string }[];
    };

export type DeleteBlockedContext = DeleteBlockedPayload & {
  entityLabel: string;
  close: () => void;
};

export interface FrontierProviderSlots {
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
