/* eslint-disable no-template-curly-in-string */
import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import * as R from 'ramda';
import Logger from '../../lib/logger';
import modifyRequest from './modifyRequest';

interface YAMLRoute {
  method: string;
  path: string;
  proxy: Record<string, string>;
  permissions?: Array<Record<string, any>>;
  hooks?: Array<Record<string, any>>;
}

const PARAM_REGEX = '{(.*?)}';
const NODE_ENV_REGEX = '\\${(.*?)}';

export const expand = (
  string: string,
  envs: NodeJS.ProcessEnv,
  regexString = NODE_ENV_REGEX,
  stripFromCharIndex = 2
) => {
  // eslint-disable-next-line no-template-curly-in-string
  const valuesRegex = new RegExp(regexString, 'g');
  return string.replace(valuesRegex, (matched: string) => {
    const varName = matched.substring(stripFromCharIndex, matched.length - 1);
    // expand to empty if varName not existent
    // alternatively, we can leave the ${} untouched - replace '' with matched.

    // varName format "ENVIRONMENT_NAME:defaultvalue"
    const [environmentName, ...rest] = varName.split(':');
    const envVariable = envs[environmentName];
    return envVariable !== undefined ? envVariable : rest.join(':');
  });
};

const onResponse = (extraOptions: any = {}) => async (
  err: any,
  res: any,
  request: Hapi.Request,
  h: Hapi.ResponseToolkit
) => {
  const { statusCode } = res;

  if (statusCode >= 200 && statusCode <= 299) {
    try {
      const payload = await Wreck.read(res, {
        json: 'force',
        gunzip: true
      });

      // only return the following key from the response
      const responseKeyToReturn = R.pathOr(
        '',
        ['responseKeyToReturn'],
        extraOptions
      );
      if (R.hasPath(responseKeyToReturn.split('.'), payload)) {
        const payloadToReturn = R.pathOr(
          {},
          responseKeyToReturn.split('.'),
          payload
        );
        return h.response(payloadToReturn).code(statusCode);
      }

      return h.response(payload).code(statusCode);
      // eslint-disable-next-line no-empty
    } catch (e) {
      Logger.error(`Failed to parse proxy response: ${e}`);
    }
  }

  return h.response(res).code(statusCode);
};

export const generateRoutes = (contents: Array<YAMLRoute> = []) => {
  return contents.map((route: YAMLRoute) => {
    const proxy = {
      ...(route?.proxy?.extraOptions && {
        onResponse: onResponse(route?.proxy?.extraOptions)
      }),
      async mapUri(request: Hapi.Request) {
        const { uri, protocol, host, port, path } = route.proxy;
        const queryParams = request.url.search || '';
        const proxyURI = uri
          ? `${expand(uri, request.params, PARAM_REGEX, 1)}${queryParams}`
          : `${protocol}://${host}:${port}/${expand(
              path,
              request.params,
              PARAM_REGEX,
              1
            )}${queryParams}'`;
        return {
          uri: proxyURI
        };
      },
      passThrough: true
    };

    return {
      method: route.method,
      path: route.path,
      handler(request: Hapi.Request, h: Hapi.ResponseToolkit) {
        return h.proxy(proxy);
      },
      options: {
        ext: {
          onPreHandler: {
            method: modifyRequest
          }
        },
        tags: ['api'],
        ...(!['GET', 'HEAD'].includes(route.method) && {
          payload: { parse: false }
        }),
        app: {
          iam: {
            permissions: route?.permissions || [],
            hooks: route?.hooks || []
          }
        }
      }
    };
  });
};
