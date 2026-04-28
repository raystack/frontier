import type { FrontierClientOptions } from '@raystack/frontier/client';

const config: FrontierClientOptions = {
  endpoint: '/api',
  billing: {
    basePlan: {
      title: 'Standard Plan'
    }
  },
  customization: {
    terminology: {
      appName: 'Client Demo',
      organization: { singular: 'Workspace', plural: 'Workspaces' }
    }
  }
};

export default config;
