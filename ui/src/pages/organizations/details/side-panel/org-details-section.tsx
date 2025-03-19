import { V1Beta1Organization } from "@raystack/frontier";
import { CalendarIcon } from "@radix-ui/react-icons";
import { Flex, List, Text, CopyButton, Tooltip } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import dayjs from "dayjs";
import { useContext } from "react";
import { AppContext } from "~/contexts/App";

interface OrganizationDetailsSectionProps {
  organization: V1Beta1Organization;
}

type Metadata = Record<string, any>;

export const OrganizationDetailsSection = ({
  organization,
}: OrganizationDetailsSectionProps) => {
  const { config } = useContext(AppContext);

  return (
    <List.Root>
      <List.Header>Organization Details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          URL
        </List.Label>
        <List.Value>
          <Text>
            {config?.app_url}/{organization.name}
          </Text>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Organization ID
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CopyButton
              text={organization.id || ""}
              data-test-id="copy-button"
            />
            <Tooltip message={organization.id}>
              <Text className={styles["org-details-section-org-id"]}>
                {organization.id}
              </Text>
            </Tooltip>
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
      {/* <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Created by
        </List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
         */}
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Organization size
        </List.Label>
        <List.Value>
          <List.Value>
            <Text>{(organization?.metadata as Metadata)?.["size"] || "-"}</Text>
          </List.Value>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Industry
        </List.Label>
        <List.Value>
          <Text>{(organization?.metadata as Metadata)?.["type"] || "-"}</Text>
        </List.Value>
      </List.Item>

      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Country
        </List.Label>
        <List.Value>
          <Text>
            {(organization?.metadata as Metadata)?.["country"] || "-"}
          </Text>
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
