import React from 'react';

export interface FrontierClientOptions {
  endpoint?: string;
  redirectSignup?: string;
  redirectLogin?: string;
  redirectMagicLinkVerify?: string;
  callbackUrl?: string;
  billingSupportEmail?: string;
}

export interface InitialState {
  sessionId?: string | null;
}

export interface FrontierProviderProps {
  config: FrontierClientOptions;
  children: React.ReactNode;
  initialState?: InitialState;
}
