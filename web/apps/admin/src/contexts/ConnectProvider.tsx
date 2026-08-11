import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import type { ReactNode } from "react";
import { TransportProvider } from "@connectrpc/connect-query";
import { jsonTransport as transport } from "~/connect/transport";

/*
 * staleTime 0 + refetchOnMount refetches on every mount, so navigating
 * re-requested roles, plans and products each time. Mutations invalidate their
 * own keys, and the search tables opt out with an explicit staleTime: 0.
 */
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
