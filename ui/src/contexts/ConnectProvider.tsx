import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TransportProvider } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ReactNode } from "react";

const frontierConnectEndpoint = process.env.NEXT_PUBLIC_FRONTIER_CONNECT_URL || "/frontier-connect";

// Create the transport for Connect RPC
const transport = createConnectTransport({
  baseUrl: frontierConnectEndpoint,
  useBinaryFormat: false,
  interceptors: [],
});

// Create a QueryClient instance
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

interface ConnectProviderProps {
  children: ReactNode;
}

export function ConnectProvider({ children }: ConnectProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>
        {children}
      </TransportProvider>
    </QueryClientProvider>
  );
}

export { queryClient };