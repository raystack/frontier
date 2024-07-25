import { FrontierClientOptions } from '../../../core/react/dist';

const config: FrontierClientOptions = {
  endpoint: '/api',
  billing: {
    basePlan: {
      title: 'Standard Plan'
    }
  }
};

export default config;
