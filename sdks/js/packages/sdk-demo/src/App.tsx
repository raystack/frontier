import config from '@/config/frontier';
import AuthContextProvider from '@/contexts/auth/provider';
import { customFetch } from '@/utils/custom-fetch';
import { FrontierProvider } from '@raystack/frontier/react';
import Router from './Router';

function App() {
  return (
    <FrontierProvider config={config} customFetch={customFetch}>
      <AuthContextProvider>
        <Router />
      </AuthContextProvider>
    </FrontierProvider>
  );
}

export default App;
