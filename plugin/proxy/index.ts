import Hapi from '@hapi/hapi';
import { Parser } from '../../lib/parser';
import YMLParser from '../../lib/parser/YMLParser';
import { expand, generateRoutes } from './utils';

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
    server.route(routes);
  }
};
