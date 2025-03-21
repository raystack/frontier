import {
  Button,
  Flex,
  IconButton,
  Separator,
  Text,
  Tooltip,
} from "@raystack/apsara/v1";

import styles from "./security.module.css";
import { PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1Domain } from "~/api/frontier";
import { api } from "~/api";
import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import { BlockOrganizationSection } from "./block-organization";
import { DomainsList } from "./domains-list";

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

  const fetchDomains = useCallback(async () => {
    if (!organizationId) return;
    try {
      setIsDomainLoading(true);
      const response =
        await api?.frontierServiceListOrganizationDomains(organizationId);
      const data = response?.data?.domains || [];
      setDomains(data);
    } catch (error) {
      console.error("Error fetching domains:", error);
    } finally {
      setIsDomainLoading(false);
    }
  }, [organizationId]);

  useEffect(() => {
    fetchDomains();
  }, [fetchDomains]);

  return (
    <Flex justify="center" className={styles["container"]}>
      <Flex className={styles["content"]} direction="column" gap={9}>
        <AddDomainSection />
        <DomainsList
          isLoading={isDomainLoading}
          domains={domains}
          fetchDomains={fetchDomains}
        />
        <Separator />
        <BlockOrganizationSection />
      </Flex>
    </Flex>
  );
};
