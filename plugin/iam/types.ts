import Hapi from '@hapi/hapi';

export type RequestToIAMTransformConfig = {
  requestKey: string;
  iamKey: string;
};

export type RequestKeysForIAM = 'query' | 'params' | 'payload' | 'response';

export type IAMAuthorizeActionConfig = {
  baseName: string;
  operation?: string;
};

export type IAMAuthorizeAction = string | IAMAuthorizeActionConfig;

export type IAMAuthorize = {
  action: IAMAuthorizeAction;
  resource: RequestToIAMTransformConfig[];
};

export type IAMAuthorizeList = IAMAuthorize[];

export interface IAMRouteOptionsApp extends Hapi.RouteOptionsApp {
  iam?: {
    authorize?: IAMAuthorizeList;
    manage?: {
      upsert?: {
        [index: number]: {
          resource: RequestToIAMTransformConfig[];
          resourceAttributes: RequestToIAMTransformConfig[];
        };
      };
    };
  };
}

export interface IAMRouteOptions extends Hapi.RouteOptions {
  app?: IAMRouteOptionsApp;
}

export interface IAMRoute extends Hapi.ServerRoute {
  options?: IAMRouteOptions;
}
