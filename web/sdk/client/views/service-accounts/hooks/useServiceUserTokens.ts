import {
  useQuery,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import { useMemo } from 'react';
import { useQueryClient, type QueryClient } from '@tanstack/react-query';
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

type FreshTokensMap = Record<string, ServiceUserToken>;

const getFreshTokensCacheKey = (serviceUserId: string) =>
  ['frontier:fresh-service-user-tokens', serviceUserId] as const;

export const cacheFreshServiceUserToken = (
  queryClient: QueryClient,
  serviceUserId: string,
  token: ServiceUserToken
) => {
  if (!token.id) return;
  const key = getFreshTokensCacheKey(serviceUserId);
  const prev = queryClient.getQueryData<FreshTokensMap>(key) ?? {};
  queryClient.setQueryData(key, { ...prev, [token.id]: token });
};

export const useServiceUserTokens = ({
  id,
  orgId,
  enableFetch
}: UseServiceUserTokensOptions) => {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { data: tokensData, isLoading } = useQuery(
    FrontierServiceQueries.listServiceUserTokens,
    create(ListServiceUserTokensRequestSchema, {
      id,
      orgId
    }),
    {
      enabled: Boolean(id) && Boolean(orgId) && Boolean(enableFetch)
    }
  );

  const freshTokensMap = queryClient.getQueryData<FreshTokensMap>(
    getFreshTokensCacheKey(id)
  );

  const tokens = useMemo(() => {
    const list = tokensData?.tokens ?? [];
    if (!freshTokensMap) return list;
    return list.map(t => {
      if (!t?.id) return t;
      const fresh = freshTokensMap[t.id];
      return fresh?.token ? { ...t, token: fresh.token } : t;
    });
  }, [tokensData, freshTokensMap]);

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
    cacheFreshServiceUserToken(queryClient, id, token);
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

    const freshKey = getFreshTokensCacheKey(id);
    const prevFresh = queryClient.getQueryData<FreshTokensMap>(freshKey);
    if (prevFresh && tokenId in prevFresh) {
      const { [tokenId]: _removed, ...rest } = prevFresh;
      queryClient.setQueryData(freshKey, rest);
    }
  };

  const clearFreshTokens = () => {
    queryClient.removeQueries({ queryKey: getFreshTokensCacheKey(id) });
  };

  return {
    tokens,
    isLoading,
    addToken,
    removeToken,
    clearFreshTokens
  };
};
