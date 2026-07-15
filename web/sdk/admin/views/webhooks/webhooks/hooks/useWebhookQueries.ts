import { useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";

export function useWebhookQueries() {
  const listWebhooks = useQuery(
    AdminServiceQueries.listWebhooks,
    {},
    {
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

  return {
    listWebhooks,
  };
}
