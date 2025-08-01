import type { FrontierClientOptions } from '@raystack/frontier/react';

const config: FrontierClientOptions = {
  endpoint: '/api',
  billing: {
    basePlan: {
      title: 'Standard Plan'
    }
  },
  customization: {
    terminology: {
      appName: 'SDK Demo',
      organization: { singular: 'Workspace', plural: 'Workspaces' }
    }
  }
};

export default config;
