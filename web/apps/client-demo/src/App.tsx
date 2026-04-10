import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { FrontierProvider } from '@raystack/frontier/react';
import Router from './Router';
import { v4 as uuid } from 'uuid';
import './styles.css';
import '@raystack/apsara/normalize.css';

const customHeaders = {
  'X-Request-ID': () => `client-demo:${uuid()}`
};

function App() {
  return (
    <FrontierProvider
      config={config}
      customHeaders={customHeaders}
    >
      <AuthContextProvider>
        <Router />
      </AuthContextProvider>
    </FrontierProvider>
  );
}

export default App;
