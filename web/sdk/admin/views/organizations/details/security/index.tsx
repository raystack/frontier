import {
  Button,
  Flex,
  IconButton,
  Separator,
  Text,
  Tooltip,
  toast,
} from "@raystack/apsara";

import styles from "./security.module.css";
import { PlusIcon } from "@radix-ui/react-icons";
import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import { BlockOrganizationSection } from "./block-organization";
import { DomainsList } from "./domains-list";
import { PageTitle } from "../../../../components/PageTitle";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListOrganizationDomainsRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useEffect } from "react";

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
  const { organization } = useOutletContext<OutletContext>();

  const { data: domains, isLoading, error } = useQuery(
    FrontierServiceQueries.listOrganizationDomains,
    create(ListOrganizationDomainsRequestSchema, {
      orgId: organization.id,
    }),
    {
      enabled: !!organization.id,
      select: (data) => data?.domains || [],
    },
  );

  useEffect(() => {
    if (error) {
      toast.error("Something went wrong", {
        description: "Unable to fetch domains",
      });
      console.error("Unable to fetch domains:", error);
    }
  }, [error]);

  const title = `Security | ${organization.title} | Organizations`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <Flex className={styles["content"]} direction="column" gap={9}>
        <AddDomainSection />
        <DomainsList
          isLoading={isLoading}
          domains={domains ?? []}
        />
        <Separator />
        <BlockOrganizationSection />
      </Flex>
    </Flex>
  );
};
