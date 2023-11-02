import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';

export const useOrganizationDomains = () => {
  const [domains, setDomains] = useState([]);
  const [isDomainsLoading, setIsDomainsLoading] = useState(false);
  const { client, activeOrganization: organization } = useFrontier();

  const getDomains = useCallback(async () => {
    try {
      setIsDomainsLoading(true);
      if (!organization?.id) return;
      const {
        // @ts-ignore
        data: { domains = [] }
      } = await client?.frontierServiceListOrganizationDomains(
        organization?.id
      );
      setDomains(domains);
    } catch (err) {
      console.error(err);
    } finally {
      setIsDomainsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getDomains();
  }, [client, getDomains, organization?.id]);

  const updatedDomains = useMemo(
    () =>
      isDomainsLoading
        ? ([{ id: 1 }, { id: 2 }, { id: 3 }] as any)
        : domains.length
        ? domains
        : [],
    [isDomainsLoading, domains]
  );

  return {
    isFetching: isDomainsLoading,
    domains: updatedDomains,
    refetch: getDomains
  };
};
