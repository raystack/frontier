import Hapi from '@hapi/hapi';

// eslint-disable-next-line @typescript-eslint/ban-types
export type RequestToIAMTransformConfig = Record<string, any>;
export type RequestKeysForIAM =
  | 'query'
  | 'params'
  | 'payload'
  | 'response'
  | 'headers';

export type IAMAuthorizeActionConfig = {
  baseName: string;
  operation?: string;
};

export type IAMAuthorizeAction = string | IAMAuthorizeActionConfig;

export type IAMAuthorize = {
  action: IAMAuthorizeAction;
  attributes: RequestToIAMTransformConfig[];
};

export type IAMAuthorizeList = IAMAuthorize[];

export type IAMUpsertConfig = {
  resources: RequestToIAMTransformConfig[];
  attributes: RequestToIAMTransformConfig[];
};

export interface IAMRouteOptionsApp extends Hapi.RouteOptionsApp {
  iam?: {
    permissions?: IAMAuthorizeList;
    hooks?: IAMUpsertConfig[];
  };
}

export interface IAMRouteOptions extends Hapi.RouteOptions {
  app?: IAMRouteOptionsApp;
}

export interface IAMRoute extends Hapi.ServerRoute {
  options?: IAMRouteOptions;
}
