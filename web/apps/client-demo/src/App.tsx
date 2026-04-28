import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { FrontierProvider } from '@raystack/frontier/client';
import { Toast } from '@raystack/apsara';
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
      <Toast.Provider>
        <AuthContextProvider>
          <Router />
        </AuthContextProvider>
      </Toast.Provider>
    </FrontierProvider>
  );
}

export default App;
