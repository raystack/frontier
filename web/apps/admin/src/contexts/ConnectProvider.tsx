import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import type { ReactNode } from "react";
import { TransportProvider } from "@connectrpc/connect-query";
import { jsonTransport as transport } from "~/connect/transport";

/* Otherwise every mount refetches. Mutations invalidate their own keys;
 * the search tables opt out with staleTime: 0. */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
      staleTime: 30 * 1000,
    },
  },
});

interface ConnectProviderProps {
  children: ReactNode;
}

export function ConnectProvider({ children }: ConnectProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>{children}</TransportProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

export { queryClient };
