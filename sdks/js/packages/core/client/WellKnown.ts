/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

import { RpcStatus, V1Beta1GetJWKsResponse } from './data-contracts';
import { HttpClient, RequestParams } from './http-client';

export class WellKnown<SecurityDataType = unknown> extends HttpClient<SecurityDataType> {
  /**
   * No description
   *
   * @tags Authz
   * @name FrontierServiceGetJwKs2
   * @summary Get well known JWKs
   * @request GET:/.well-known/jwks.json
   * @secure
   */
  frontierServiceGetJwKs2 = (params: RequestParams = {}) =>
    this.request<V1Beta1GetJWKsResponse, RpcStatus>({
      path: `/.well-known/jwks.json`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
}
