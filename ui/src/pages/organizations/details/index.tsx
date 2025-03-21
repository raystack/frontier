import { V1Beta1Organization } from "~/api/frontier";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { Outlet, useParams } from "react-router-dom";

import { OrganizationDetailsLayout } from "./layout";

export const OrganizationDetails = () => {
  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(true);
  const { organizationId } = useParams();

  async function fetchOrganization(id: string) {
    try {
      const response = await api?.frontierServiceGetOrganization(id);
      const org = response.data?.organization;
      setOrganization(org);
    } catch (error) {
      console.error(error);
    } finally {
      setIsOrganizationLoading(false);
    }
  }

  useEffect(() => {
    if (organizationId) {
      fetchOrganization(organizationId);
    }
  }, [organizationId]);

  return (
    <OrganizationDetailsLayout
      organization={organization}
      isLoading={isOrganizationLoading}
    >
      <Outlet
        context={{
          organizationId: organization?.id,
          fetchOrganization,
          organization,
        }}
      />
    </OrganizationDetailsLayout>
  );
};
