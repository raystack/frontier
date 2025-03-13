import { Flex, List, Separator, Text } from "@raystack/apsara/v1";
import styles from "./details.module.css";
import { V1Beta1Organization, V1Beta1OrganizationKyc } from "~/api/frontier";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { api } from "~/api";
import Skeleton from "react-loading-skeleton";
import CheckCircleFilledIcon from "~/assets/icons/check-circle-filled.svg?react";
import CrossCircleFilledIcon from "~/assets/icons/cross-circle-filled.svg?react";

const KYCDetails = ({ organizationId }: { organizationId: string }) => {
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
        // Update state with fetched data
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
        <List.Label minWidth="88px">Status</List.Label>
        <List.Value>
          {isKYCLoading ? (
            <Skeleton />
          ) : KYCDetails?.status ? (
            <Flex justifyContent="center" alignItems="center" gap={3}>
              <CheckCircleFilledIcon
                color={"var(--rs-color-foreground-success-primary)"}
              />
              <Text>Verified</Text>
            </Flex>
          ) : (
            <Flex justifyContent="center" alignItems="center" gap={3}>
              <CrossCircleFilledIcon
                color={"var(--rs-color-foreground-danger-primary)"}
              />
              <Text>Not verified</Text>
            </Flex>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Documents Link</List.Label>
        <List.Value>
          {isKYCLoading ? <Skeleton /> : KYCDetails?.link || "N/A"}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};

interface OrganizationDetailsProps {
  organization: V1Beta1Organization;
}

const OrganizationDetails = ({ organization }: OrganizationDetailsProps) => {
  return (
    <List.Root>
      <List.Header>Organization Details</List.Header>
      <List.Item>
        <List.Label minWidth="88px">URL</List.Label>
        <List.Value>{organization.name}</List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Organization ID</List.Label>
        <List.Value>{organization.id}</List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Created on</List.Label>
        <List.Value>
          {dayjs(organization?.created_at).format("DD MMM YYYY")}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Created by</List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Organization size</List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Industry</List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Country</List.Label>
        <List.Value>
          <span />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="88px">Status</List.Label>
        <List.Value>{organization?.state}</List.Value>
      </List.Item>
    </List.Root>
  );
};

interface SidePanelProps {
  organization: V1Beta1Organization;
}

export function SidePanel({ organization }: SidePanelProps) {
  return (
    <aside data-test-id="admin-ui-sidepanel" className={styles["side-panel"]}>
      <OrganizationDetails organization={organization} />
      <Separator />
      <KYCDetails organizationId={organization.id || ""} />
      <Separator />
      <List.Root>
        <List.Header>Plan details</List.Header>
        <List.Item>
          <List.Label minWidth="88px">Name</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Started from</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Ends on</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Payment mode</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
      </List.Root>
      <Separator />
      <List.Root>
        <List.Header>Plan details</List.Header>
        <List.Item>
          <List.Label minWidth="88px">Name</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Started from</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Ends on</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
        <List.Item>
          <List.Label minWidth="88px">Payment mode</List.Label>
          <List.Value>Active</List.Value>
        </List.Item>
      </List.Root>
    </aside>
  );
}
