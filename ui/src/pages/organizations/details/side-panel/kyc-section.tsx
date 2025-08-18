import { Flex, List, Text, Link, Tooltip } from "@raystack/apsara";
import { useContext } from "react";
import styles from "./side-panel.module.css";
import Skeleton from "react-loading-skeleton";
import {
  CheckCircleFilledIcon,
  CrossCircleFilledIcon,
} from "@raystack/apsara/icons";
import { Link2Icon } from "@radix-ui/react-icons";
import { OrganizationContext } from "../contexts/organization-context";

export const KYCDetailsSection = () => {
  const { isKYCLoading, kycDetails } = useContext(OrganizationContext);
  return (
    <List.Root>
      <List.Header>KYC Details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Status
        </List.Label>
        <List.Value>
          {isKYCLoading ? (
            <Skeleton />
          ) : kycDetails?.status ? (
            <Flex justify="start" align="center" gap={3}>
              <CheckCircleFilledIcon
                color={"var(--rs-color-foreground-success-primary)"}
                className={styles["kyc-status-icon"]}
              />
              <Text>Verified</Text>
            </Flex>
          ) : (
            <Flex justify="start" align="center" gap={3}>
              <CrossCircleFilledIcon
                color={"var(--rs-color-foreground-danger-primary)"}
                className={styles["kyc-status-icon"]}
              />
              <Text>Not verified</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Documents Link
        </List.Label>
        <List.Value>
          {isKYCLoading ? (
            <Skeleton />
          ) : (
            <Flex justify="start" align="center" gap={3}>
              {kycDetails?.link ? (
                <>
                  <Link2Icon />
                  <Tooltip message={kycDetails?.link}>
                    <Link
                      href={kycDetails?.link}
                      target="_blank"
                      data-test-id="kyc-link"
                      className={styles["kyc_link"]}
                    >
                      {kycDetails?.link}
                    </Link>
                  </Tooltip>
                </>
              ) : (
                <Text>N/A</Text>
              )}
            </Flex>
          )}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
