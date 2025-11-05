import {
  Button,
  Flex,
  IconButton,
  Separator,
  Text,
  Tooltip,
} from "@raystack/apsara";

import styles from "./security.module.css";
import { PlusIcon } from "@radix-ui/react-icons";
import { useOutletContext } from "react-router-dom";
import { OutletContext } from "../types";
import { BlockOrganizationSection } from "./block-organization";
import { DomainsList } from "./domains-list";
import PageTitle from "~/components/page-title";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, ListOrganizationDomainsRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

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
  const { organizationId, organization } = useOutletContext<OutletContext>();

  const { data: domainsData, isLoading } = useQuery(
    FrontierServiceQueries.listOrganizationDomains,
    create(ListOrganizationDomainsRequestSchema, {
      orgId: organizationId,
    }),
    {
      enabled: !!organizationId,
    },
  );

  const domains = domainsData?.domains || [];
  const title = `Security | ${organization.title} | Organizations`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <Flex className={styles["content"]} direction="column" gap={9}>
        <AddDomainSection />
        <DomainsList
          isLoading={isLoading}
          domains={domains}
          organizationId={organizationId}
        />
        <Separator />
        <BlockOrganizationSection />
      </Flex>
    </Flex>
  );
};
