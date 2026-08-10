import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import type { ReactNode } from "react";
import { TransportProvider } from "@connectrpc/connect-query";
import { jsonTransport as transport } from "~/connect/transport";

/*
 * staleTime defaults to 0, which combined with refetchOnMount means every
 * mount of every component refetches — reference data like roles, plans and
 * products was re-requested on each navigation. A short window covers
 * navigating between pages without holding data long enough to look stale;
 * mutations invalidate their own keys, so writes are still reflected at once.
 * The search-backed tables opt out with an explicit staleTime: 0.
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
