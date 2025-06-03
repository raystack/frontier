import type { FrontierClientOptions } from '@raystack/frontier/react';

const config: FrontierClientOptions = {
  endpoint: '/api',
  billing: {
    basePlan: {
      title: 'Standard Plan'
    }
  }
};

export default config;
