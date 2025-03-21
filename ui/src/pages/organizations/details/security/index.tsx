import {
  Button,
  Flex,
  IconButton,
  Separator,
  Text,
  Tooltip,
} from "@raystack/apsara/v1";

import styles from "./security.module.css";
import { CheckCircledIcon, PlusIcon, TrashIcon } from "@radix-ui/react-icons";
import { useEffect, useState } from "react";
import { V1Beta1Domain } from "~/api/frontier";
import { api } from "~/api";
import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import Skeleton from "react-loading-skeleton";
import { BlockOrganizationSection } from "./block-organization";

interface DomainItemProps {
  domain: V1Beta1Domain;
}

const DomainItem = ({ domain }: DomainItemProps) => {
  return (
    <Flex className={styles["domains-list-item"]} justify="between">
      <Flex gap={3}>
        <Text size={3}>{domain?.name}</Text>
        {domain.state === "verified" ? (
          <CheckCircledIcon
            color={"var(--rs-color-foreground-success-primary)"}
          />
        ) : null}
      </Flex>
      <IconButton size={3} data-test-id="delete-domain-button">
        <TrashIcon />
      </IconButton>
    </Flex>
  );
};

interface DomainListProps {
  isLoading: boolean;
  domains: V1Beta1Domain[];
}

const DomainList = ({ isLoading, domains = [] }: DomainListProps) => {
  return isLoading ? (
    <Flex direction="column" className={styles["domains-list"]}>
      {[...Array(3)].map((_, index) => (
        <Skeleton
          key={index}
          height={20}
          style={{ margin: "var(--rs-space-5) 0" }}
        />
      ))}
    </Flex>
  ) : domains.length ? (
    <Flex direction="column" className={styles["domains-list"]}>
      {domains.map((domain) => (
        <DomainItem key={domain?.id} domain={domain} />
      ))}
    </Flex>
  ) : null;
};

const AddDomainSection = () => {
  return (
    <Flex gap={5} justify="between">
      <Flex direction="column" gap={3}>
        <Text size={5}>Allowed email domains</Text>
        <Text size={3} variant="secondary">
          Anyone with an email address at these domains is allowed to sign up
          for this workspace.
        </Text>
      </Flex>
      <Tooltip message="Coming soon">
        <Button
          variant="outline"
          color="neutral"
          leadingIcon={<PlusIcon />}
          size="small"
          data-test-id="add-domain-button"
          disabled={true}
        >
          Add domain
        </Button>
      </Tooltip>
    </Flex>
  );
};

export const OrganizationSecurity = () => {
  const { organizationId } = useOutletContext<OutletContext>();
  const [domains, setDomains] = useState<V1Beta1Domain[]>([]);
  const [isDomainLoading, setIsDomainLoading] = useState(false);

  async function fetchDomains(orgId: string) {
    try {
      setIsDomainLoading(true);
      const response = await api?.frontierServiceListOrganizationDomains(orgId);
      const data = response?.data?.domains || [];
      setDomains(data);
    } catch (error) {
      console.error("Error fetching domains:", error);
    } finally {
      setIsDomainLoading(false);
    }
  }

  useEffect(() => {
    if (organizationId) fetchDomains(organizationId);
  }, [organizationId]);

  return (
    <Flex justify="center" className={styles["container"]}>
      <Flex className={styles["content"]} direction="column" gap={9}>
        <AddDomainSection />
        <DomainList isLoading={isDomainLoading} domains={domains} />
        <Separator />
        <BlockOrganizationSection />
      </Flex>
    </Flex>
  );
};
