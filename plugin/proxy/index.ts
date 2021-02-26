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
    const ROUTES_WITH_PROXY_DATA = ROUTES.map(
      R.assocPath(
        ['handler', 'proxy', 'httpClient', 'request'],
        httpRequestClient
      )
    );

    server.route(ROUTES_WITH_PROXY_DATA);
  }
};
