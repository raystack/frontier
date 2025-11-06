import {
  useQuery,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListServiceUserTokensRequestSchema,
  ListServiceUserTokensResponseSchema,
  type ServiceUserToken
} from '@raystack/proton/frontier';

interface UseServiceUserTokensOptions {
  id: string;
  orgId: string;
  enableFetch?: boolean;
}

export const useServiceUserTokens = ({
  id,
  orgId,
  enableFetch
}: UseServiceUserTokensOptions) => {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { data: tokens = [], isLoading } = useQuery(
    FrontierServiceQueries.listServiceUserTokens,
    create(ListServiceUserTokensRequestSchema, {
      id,
      orgId
    }),
    {
      enabled: Boolean(id) && Boolean(orgId) && Boolean(enableFetch),
      select: data => data?.tokens ?? []
    }
  );

  const getQueryKey = () => {
    return createConnectQueryKey({
      schema: FrontierServiceQueries.listServiceUserTokens,
      transport,
      input: create(ListServiceUserTokensRequestSchema, {
        id,
        orgId
      }),
      cardinality: 'finite'
    });
  };

  const addToken = (token: ServiceUserToken) => {
    const queryKey = getQueryKey();
    const currentData = queryClient.getQueryData(queryKey);
    const existingTokens = currentData?.tokens ?? [];
    const filteredTokens = existingTokens.filter(t => t !== undefined);
    queryClient.setQueryData(
      queryKey,
      create(ListServiceUserTokensResponseSchema, {
        tokens: [token, ...filteredTokens]
      })
    );
  };

  const removeToken = (tokenId: string) => {
    const queryKey = getQueryKey();
    const currentData = queryClient.getQueryData(queryKey);
    const existingTokens = currentData?.tokens ?? [];
    const filteredTokens = existingTokens.filter(
      t => t?.id !== tokenId && t !== undefined
    );
    queryClient.setQueryData(
      queryKey,
      create(ListServiceUserTokensResponseSchema, {
        tokens: filteredTokens
      })
    );
  };

  return {
    tokens,
    isLoading,
    addToken,
    removeToken
  };
};
