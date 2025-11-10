import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import { FrontierServiceQueries, GetOrganizationDomainRequestSchema, type Domain } from '@raystack/proton/frontier';
import { useFrontier } from '../contexts/FrontierContext';

export interface UseOrganizationDomainReturn {
  domain: Domain | undefined;
  isLoading: boolean;
  error: unknown;
}

export const useOrganizationDomain = (domainId?: string): UseOrganizationDomainReturn => {
  const { activeOrganization: organization } = useFrontier();

  const {
    data: domain,
    isLoading,
    error: domainError
  } = useQuery(
    FrontierServiceQueries.getOrganizationDomain,
    create(GetOrganizationDomainRequestSchema, {
      id: domainId || '',
      orgId: organization?.id || ''
    }),
    {
      enabled: !!domainId && !!organization?.id,
      select: (d) => d?.domain
    }
  );

  return {
    domain,
    isLoading,
    error: domainError
  };
};

