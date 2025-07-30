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
      appName: 'SDK Demo'
    }
  }
};

export default config;
