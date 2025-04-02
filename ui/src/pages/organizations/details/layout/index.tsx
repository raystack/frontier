import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./layout.module.css";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { OrgSidePanel } from "../side-panel/";
import { V1Beta1Organization } from "~/api/frontier";
import React, { useState } from "react";
import LoadingState from "~/components/states/Loading";
import { OrganizationIcon } from "@raystack/apsara/icons";
import PageTitle from "~/components/page-title";

interface OrganizationDetailsLayoutProps {
  isLoading: boolean;
  organization?: V1Beta1Organization;
  children: React.ReactNode;
}

export const OrganizationDetailsLayout = ({
  isLoading,
  organization,
  children,
}: OrganizationDetailsLayoutProps) => {
  const [showSidePanel, setShowSidePanel] = useState(true);

  function toggleSidePanel() {
    setShowSidePanel(!showSidePanel);
  }

  const title = `${organization?.title} | Organizations`;

  return isLoading ? (
    // TODO: make better loading state for page
    <LoadingState />
  ) : organization ? (
    <Flex direction="column" className={styles.page}>
      <PageTitle title={title} />
      <OrganizationsDetailsNavabar
        organization={organization}
        toggleSidePanel={toggleSidePanel}
      />
      <Flex justify="between" style={{ height: "100%" }}>
        <Flex
          className={
            showSidePanel
              ? styles["main_content_with_sidepanel"]
              : styles["main_content"]
          }
        >
          {children}
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
