/* eslint-disable no-template-curly-in-string */
import Hapi from '@hapi/hapi';

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

export const generateRoutes = (contents: Array<YAMLRoute> = []) => {
  return contents.map((route: YAMLRoute) => {
    return {
      method: route.method,
      path: route.path,
      handler: {
        proxy: {
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
        }
      },
      options: {
        tags: ['api'],
        app: {
          iam: {
            permissions: route?.permissions || [],
            hooks: route?.hooks || []
          },
          proxy: route?.proxy?.extraOptions
        }
      }
    };
  });
};
