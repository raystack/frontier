import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./details.module.css";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { OrgSidePanel } from "./side-panel/";
import { V1Beta1Organization } from "~/api/frontier";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { useParams } from "react-router-dom";
import LoadingState from "~/components/states/Loading";
import { OrganizationIcon } from "@raystack/apsara/icons";
import PageTitle from "~/components/page-title";

export const OrganizationDetails = () => {
  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(true);
  const { organizationId } = useParams();
  const [showSidePanel, setShowSidePanel] = useState(true);

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

  function toggleSidePanel() {
    setShowSidePanel(!showSidePanel);
  }

  return isOrganizationLoading ? (
    // TODO: make better loading state for page
    <LoadingState />
  ) : organization ? (
    <Flex direction="column" className={styles.page}>
      <PageTitle title={organization?.title} />
      <OrganizationsDetailsNavabar
        organization={organization}
        toggleSidePanel={toggleSidePanel}
      />
      <Flex justify="between" style={{ height: "100%" }}>
        <Flex style={{ width: "100%" }}>
          <EmptyState icon={<OrganizationIcon />} heading="Coming Soon" />
        </Flex>
        {showSidePanel ? <OrgSidePanel organization={organization} /> : null}
      </Flex>
    </Flex>
  ) : (
    <Flex
      style={{ height: "100vh", width: "100%" }}
      align="center"
      justify="center"
    >
      <PageTitle title={"Organization not found"} />
      <EmptyState
        icon={<OrganizationIcon />}
        heading="Organization not found"
        subHeading="The organization you are looking for does not exist."
      />
    </Flex>
  );
};
