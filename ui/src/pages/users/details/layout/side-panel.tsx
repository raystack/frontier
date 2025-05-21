import { useEffect, useCallback, useState } from "react";
import { Avatar, getAvatarColor, SidePanel } from "@raystack/apsara/v1";
import { SidePanelDetails } from "./side-panel-details";
import { SidePanelMembership } from "./side-panel-membership";
import styles from "./side-panel.module.css";
import { getUserName } from "../../util";
import { useUser } from "../user-context";
import { SearchUserOrganizationsResponseUserOrganization } from "~/api/frontier";
import { api } from "~/api";

export const UserDetailsSidePanel = () => {
  const user = useUser();

  const [userOrganizations, setUserOrganizations] =
    useState<SearchUserOrganizationsResponseUserOrganization[]>();
  const [isLoading, setIsLoading] = useState(true);

  const fetchUserOrgs = useCallback(async (id: string) => {
    try {
      setIsLoading(true);
      const response = await api?.adminServiceSearchUserOrganizations(id, {});
      setUserOrganizations(response.data?.user_organizations);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const onReset = () => {
    if (user?.id) fetchUserOrgs(user.id);
  };

  useEffect(onReset, [user, fetchUserOrgs]);

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
      {isLoading ? (
        <SidePanel.Section>
          <SidePanelMembership showTitle isLoading />
        </SidePanel.Section>
      ) : (
        userOrganizations?.map((org, index) => (
          <SidePanel.Section key={org.org_id}>
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
