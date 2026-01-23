import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { customFetch } from '@/utils/custom-fetch';
import { FrontierProvider } from '@raystack/frontier/react';
import Router from './Router';
import { v4 as uuid } from 'uuid';

const customHeaders = {
  'X-Request-ID': () => `client-demo:${uuid()}`
};

function App() {
  return (
    <FrontierProvider
      config={config}
      customFetch={customFetch}
      customHeaders={customHeaders}
    >
      <AuthContextProvider>
        <Router />
      </AuthContextProvider>
    </FrontierProvider>
  );
}

export default App;
