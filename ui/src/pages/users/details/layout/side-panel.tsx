import { Avatar, getAvatarColor, SidePanel, Text } from "@raystack/apsara";
import { SidePanelDetails } from "./side-panel-details";
import { SidePanelMembership } from "./side-panel-membership";
import styles from "./side-panel.module.css";
import { getUserName } from "../../util";
import { useUser } from "../user-context";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useQuery } from "@connectrpc/connect-query";

export const UserDetailsSidePanel = () => {
  const { user } = useUser();

  const {
    data: userOrganizationsResponse,
    isLoading,
    refetch: onReset,
    error,
  } = useQuery(
    AdminServiceQueries.searchUserOrganizations,
    {
      id: user?.id || "",
      query: {},
    },
    {
      enabled: !!user?.id,
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

  const userOrganizations = userOrganizationsResponse?.userOrganizations || [];

  return (
    <SidePanel
      data-test-id="admin-ui-user-details-sidepanel"
      className={styles["side-panel"]}>
      <SidePanel.Header
        title={getUserName(user)}
        icon={
          <Avatar
            fallback={getUserName(user)?.[0]}
            color={getAvatarColor(user?.id || "")}
            src={user?.avatar}
          />
        }
      />
      <SidePanel.Section>
        <SidePanelDetails />
      </SidePanel.Section>
      {error ? (
        <SidePanel.Section>
          <Text variant="danger">Failed to load user organizations</Text>
        </SidePanel.Section>
      ) : isLoading ? (
        <SidePanel.Section>
          <SidePanelMembership showTitle isLoading />
        </SidePanel.Section>
      ) : (
        userOrganizations?.map((org, index) => (
          <SidePanel.Section key={org.orgId}>
            <SidePanelMembership
              data={org}
              showTitle={index === 0}
              onReset={onReset}
            />
          </SidePanel.Section>
        ))
      )}
    </SidePanel>
  );
};
