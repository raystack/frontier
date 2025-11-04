import { useFrontier } from '../contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, ListOrganizationDomainsRequestSchema, type Domain } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export interface UseOrganizationDomainsReturn {
  isFetching: boolean;
  domains: Domain[];
  refetch: () => void;
  error: unknown;
}

export const useOrganizationDomains = (): UseOrganizationDomainsReturn => {
  const { activeOrganization: organization } = useFrontier();

  const {
    data: domainsData,
    isLoading: isDomainsLoading,
    error: domainsError,
    refetch: refetchDomains
  } = useQuery(
    FrontierServiceQueries.listOrganizationDomains,
    create(ListOrganizationDomainsRequestSchema, {
      orgId: organization?.id || ''
    }),
    {
      enabled: !!organization?.id,
      select: (d) => d?.domains ?? []
    }
  );

  return {
    isFetching: isDomainsLoading,
    domains: domainsData || [],
    refetch: refetchDomains,
    error: domainsError
  };
};
