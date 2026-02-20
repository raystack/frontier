import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./layout.module.css";
import { EmptyState, Flex, Spinner } from "@raystack/apsara";
import { OrgSidePanel } from "../side-panel/";
import React, { useState } from "react";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { PageTitle } from "../../../../components/PageTitle";
import { EditKYCPanel } from "../edit/kyc";
import { EditOrganizationPanel } from "../edit/organization";
import { EditBillingPanel } from "../edit/billing";
import type { Organization } from "@raystack/proton/frontier";

interface OrganizationDetailsLayoutProps {
  isLoading: boolean;
  organization?: Organization;
  children: React.ReactNode;
  onExportMembers?: () => Promise<void>;
  onExportProjects?: () => Promise<void>;
  onExportTokens?: () => Promise<void>;
}

export const OrganizationDetailsLayout = ({
  isLoading,
  organization,
  children,
  onExportMembers,
  onExportProjects,
  onExportTokens,
}: OrganizationDetailsLayoutProps) => {
  const [showSidePanel, setShowSidePanel] = useState(true);
  const [showKYCPanel, setShowKYCPanel] = useState(false);
  const [showEditOrgPanel, setShowEditOrgPanel] = useState(false);
  const [showEditBillingPanel, setShowEditBillingPanel] = useState(false);

  function toggleSidePanel() {
    setShowSidePanel(!showSidePanel);
  }

  function closeKYCPanel() {
    setShowKYCPanel(false);
  }

  function openKYCPanel() {
    setShowKYCPanel(true);
  }

  function closeEditOrgPanel() {
    setShowEditOrgPanel(false);
  }

  function openEditOrgPanel() {
    setShowEditOrgPanel(true);
  }

  function openEditBillingPanel() {
    setShowEditBillingPanel(true);
  }

  function closeEditBillingPanel() {
    setShowEditBillingPanel(false);
  }

  const title = `${organization?.title} | Organizations`;

  return isLoading ? (
    <Flex align="center" justify="center" style={{ minHeight: "200px", width: "100%" }}>
      <Spinner size={6} />
    </Flex>
  ) : organization ? (
    <Flex direction="column" className={styles.page}>
      <PageTitle title={title} />
      <OrganizationsDetailsNavabar
        organization={organization}
        toggleSidePanel={toggleSidePanel}
        openKYCPanel={openKYCPanel}
        openEditOrgPanel={openEditOrgPanel}
        openEditBillingPanel={openEditBillingPanel}
        onExportMembers={onExportMembers}
        onExportProjects={onExportProjects}
        onExportTokens={onExportTokens}
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
        {showKYCPanel ? <EditKYCPanel onClose={closeKYCPanel} /> : null}
        {showSidePanel ? <OrgSidePanel organization={organization} /> : null}
        {showEditOrgPanel ? (
          <EditOrganizationPanel onClose={closeEditOrgPanel} />
        ) : null}
        {showEditBillingPanel ? (
          <EditBillingPanel onClose={closeEditBillingPanel} />
        ) : null}
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
