import { V1Beta1Organization } from "@raystack/frontier";
import { CalendarIcon } from "@radix-ui/react-icons";
import { Flex, List, Text } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import dayjs from "dayjs";
import { CopyButton } from "~/components/copy-button";

interface OrganizationDetailsSectionProps {
  organization: V1Beta1Organization;
}

export const OrganizationDetailsSection = ({
  organization,
}: OrganizationDetailsSectionProps) => {
  return (
    <List.Root>
      <List.Header>Organization Details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          URL
        </List.Label>
        <List.Value>{organization.name}</List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Organization ID
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CopyButton text={organization.id || ""} resetDelay={1000} />
            <Text>{organization.id}</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Created on
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CalendarIcon />
            <Text>{dayjs(organization?.created_at).format("DD MMM YYYY")}</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Created by
        </List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Organization size
        </List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Industry
        </List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Country
        </List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Status
        </List.Label>
        <List.Value>{organization?.state}</List.Value>
      </List.Item>
    </List.Root>
  );
};
