import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./details.module.css";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { SidePanel } from "./side-panel";
import { V1Beta1Organization } from "~/api/frontier";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { useParams } from "react-router-dom";
import LoadingState from "~/components/states/Loading";
import OrganizationsIcon from "~/assets/icons/organization.svg?react";

export const OrganizationDetails = () => {
  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(false);
  const { organizationId } = useParams();

  useEffect(() => {
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
    if (organizationId) {
      fetchOrganization(organizationId);
    }
  }, [organizationId]);

  return isOrganizationLoading ? (
    // TODO: make better loading state for page
    <LoadingState />
  ) : organization ? (
    <Flex direction="column" className={styles.page}>
      <OrganizationsDetailsNavabar organization={organization} />
      <Flex justify="between" style={{ height: "100%" }}>
        <p>This is the details page for an organization.</p>
        <SidePanel organization={organization} />
      </Flex>
    </Flex>
  ) : (
    <Flex
      style={{ height: "100vh", width: "100%" }}
      align="center"
      justify="center"
    >
      <EmptyState
        icon={<OrganizationsIcon />}
        heading="Organization not found"
        subHeading="The organization you are looking for does not exist."
      />
    </Flex>
  );
};
