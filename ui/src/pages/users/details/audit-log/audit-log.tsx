import { EmptyState, Flex, Button } from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import ChipIcon from "~/assets/icons/cpu-chip.svg?react";
import { useAppContext } from "~/contexts/App";
import { useUser } from "../user-context";

export const UserDetailsAuditLogPage = () => {
  const user = useUser();
  const { config } = useAppContext();

  const title = `Audit Log | ${user?.email} | Users`;

  return (
    <Flex style={{ width: "100%" }}>
      <PageTitle title={title} />
      <EmptyState
        variant="empty2"
        heading="Audit log"
        subHeading={`The audit log in ${config.title} provides a detailed record of all key actions taken within the platform. Track user activity, monitor changes, and ensure security and compliance with a transparent activity log`}
        icon={<ChipIcon />}
        primaryAction={
          <Button disabled color="neutral" variant="outline">
            Coming soon
          </Button>
        }
      />
    </Flex>
  );
};
