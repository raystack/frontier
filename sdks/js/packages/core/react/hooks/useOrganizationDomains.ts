import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';

export const useOrganizationDomains = () => {
  const [domains, setDomains] = useState<V1Beta1Domain[]>([]);
  const [isDomainsLoading, setIsDomainsLoading] = useState(false);
  const { client, activeOrganization: organization } = useFrontier();

  const getDomains = useCallback(async () => {
    try {
      setIsDomainsLoading(true);
      if (!organization?.id) return;
      const resp = await client?.frontierServiceListOrganizationDomains(
        organization?.id
      );
      const data = resp?.data?.domains || [];
      setDomains(data);
    } catch (err) {
      console.error(err);
    } finally {
      setIsDomainsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getDomains();
  }, [getDomains]);

  return {
    isFetching: isDomainsLoading,
    domains: domains,
    refetch: getDomains
  };
};
