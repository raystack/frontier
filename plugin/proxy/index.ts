import Hapi from '@hapi/hapi';
import Wreck from '@hapi/wreck';
import * as R from 'ramda';
import getProxyOptions from './proxyOptions';
import { Parser } from '../../lib/parser';
import YMLParser from '../../lib/parser/YMLParser';
import { expand, generateRoutes } from './utils';

// TODO: Replace options: any with correct type. Currently not possible because Wreck is not exposing types for some reason
const httpRequestClient = async (method: string, uri: string, options: any) => {
  const proxyOptions = await getProxyOptions(options);
  return Wreck.request(method, uri, proxyOptions);
};

const onResponse = (route: any) => async (
  err: any,
  res: any,
  request: Hapi.Request,
  h: Hapi.ResponseToolkit
) => {
  const payload = await Wreck.read(res, {
    json: 'force',
    gunzip: true
  });

  // only return the following key from the response
  const responseKeyToReturn = R.pathOr(
    '',
    ['options', 'app', 'proxy', 'responseKeyToReturn'],
    route
  );
  if (!R.hasPath(responseKeyToReturn.split('.'), payload)) {
    return h.response(payload);
  }

  const payloadToReturn = R.pathOr({}, responseKeyToReturn.split('.'), payload);
  return h.response(payloadToReturn);
};

export const plugin = {
  name: 'proxies',
  dependencies: [],
  async register(server: Hapi.Server) {
    const parser: Parser = YMLParser();
    const contents = await parser.parseFolder('proxies');

    // expand with Node environment values
    const routesContent = JSON.parse(
      expand(JSON.stringify(contents), process.env)
    );
    const routes = generateRoutes(routesContent);

    const ROUTES: any[] = routes;
    const ROUTES_WITH_PROXY_DATA: any[] = ROUTES.map((route) => {
      return R.pipe(
        R.assocPath(
          ['handler', 'proxy', 'httpClient', 'request'],
          httpRequestClient
        ),
        R.assocPath(['handler', 'proxy', 'onResponse'], onResponse(route))
      )(route);
    });

    server.route(ROUTES_WITH_PROXY_DATA);
  }
};
