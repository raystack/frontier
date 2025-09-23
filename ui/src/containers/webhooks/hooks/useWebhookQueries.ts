import { createConnectQueryKey, useTransport, useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useQueryClient } from "@tanstack/react-query";

export function useWebhookQueries() {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const listWebhooks = useQuery(
    AdminServiceQueries.listWebhooks,
    {},
    {
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

  const invalidateWebhooksList = async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: AdminServiceQueries.listWebhooks,
        transport,
        input: {},
        cardinality: "finite",
      }),
    });
  };

  return {
    listWebhooks,
    invalidateWebhooksList
  };
}
