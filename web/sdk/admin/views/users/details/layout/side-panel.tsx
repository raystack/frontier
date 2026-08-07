import { Avatar, getAvatarColor, SidePanel, Text } from "@raystack/apsara";
import { SidePanelDetails } from "./side-panel-details";
import { SidePanelMembership } from "./side-panel-membership";
import { SidePanelInvitation } from "./side-panel-invitation";
import styles from "./side-panel.module.css";
import { getUserName } from "../../util";
import { useUser } from "../user-context";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
} from "@raystack/proton/frontier";
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

  const {
    data: invitationsResponse,
    isLoading: isInvitationsLoading,
    error: invitationsError,
  } = useQuery(
    FrontierServiceQueries.listUserInvitations,
    // `id` is the user's email, not their uuid — invitations are keyed by email
    // since the invitee may not have an account yet.
    {
      id: user?.email || "",
    },
    {
      enabled: !!user?.email,
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

  const userOrganizations = userOrganizationsResponse?.userOrganizations || [];
  const invitations = invitationsResponse?.invitations || [];

  return (
    <SidePanel
      data-test-id="admin-user-details-sidepanel"
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
      {invitationsError ? (
        <SidePanel.Section>
          <Text variant="danger">Failed to load user invitations</Text>
        </SidePanel.Section>
      ) : isInvitationsLoading ? (
        <SidePanel.Section>
          <SidePanelInvitation showTitle isLoading />
        </SidePanel.Section>
      ) : (
        invitations?.map((invite, index) => (
          <SidePanel.Section key={invite.id}>
            <SidePanelInvitation data={invite} showTitle={index === 0} />
          </SidePanel.Section>
        ))
      )}
    </SidePanel>
  );
};
