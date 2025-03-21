import { Flex, List, Text } from "@raystack/apsara/v1";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { V1Beta1OrganizationKyc } from "~/api/frontier";
import styles from "./side-panel.module.css";
import Skeleton from "react-loading-skeleton";
import {
  CheckCircleFilledIcon,
  CrossCircleFilledIcon,
} from "@raystack/apsara/icons";
import { Link2Icon } from "@radix-ui/react-icons";

export const KYCDetailsSection = ({
  organizationId,
}: {
  organizationId: string;
}) => {
  const [KYCDetails, setKYCDetails] = useState<V1Beta1OrganizationKyc | null>(
    null,
  );
  const [isKYCLoading, setIsKYCLoading] = useState(true);

  useEffect(() => {
    async function fetchKYCDetails(id: string) {
      setIsKYCLoading(true);
      try {
        const response = await api?.frontierServiceGetOrganizationKyc(id);
        const kyc = response?.data?.organization_kyc || null;
        setKYCDetails(kyc);
      } catch (error) {
        console.error("Error fetching KYC details:", error);
      } finally {
        setIsKYCLoading(false);
      }
    }
    if (organizationId) {
      fetchKYCDetails(organizationId);
    }
  }, [organizationId]);

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
          ) : KYCDetails?.status ? (
            <Flex justifyContent="center" alignItems="center" gap={3}>
              <CheckCircleFilledIcon
                color={"var(--rs-color-foreground-success-primary)"}
                className={styles["kyc-status-icon"]}
              />
              <Text>Verified</Text>
            </Flex>
          ) : (
            <Flex justifyContent="center" alignItems="center" gap={3}>
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
            <Flex justifyContent="center" alignItems="center" gap={3}>
              {KYCDetails?.link ? (
                <>
                  <Link2Icon />
                  <Text>{KYCDetails?.link}</Text>
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
