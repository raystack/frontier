import Hapi from '@hapi/hapi';

export type RequestToIAMTransformObj = {
  requestKey: string;
  iamKey: string;
};

export type RequestKeysForIAM = 'query' | 'params' | 'payload' | 'response';

export type RequestToIAMTransformConfig = Partial<
  Record<RequestKeysForIAM, RequestToIAMTransformObj[]>
>;

export type IAMAuthorize = {
  action: string;
  resource: RequestToIAMTransformConfig;
};

export type IAMAuthorizeList = IAMAuthorize[];

export interface IAMRouteOptionsApp extends Hapi.RouteOptionsApp {
  iam?: {
    authorize?: IAMAuthorizeList;
    manage?: {
      upsert?: {
        [index: number]: {
          resource: RequestToIAMTransformConfig;
          resourceAttributes: RequestToIAMTransformConfig;
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
