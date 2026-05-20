import {
  Button,
  Flex,
  Separator,
  Text,
  Tooltip,
  toastManager,
} from "@raystack/apsara-v1";

import styles from './security.module.css';
import { PlusIcon } from '@radix-ui/react-icons';
import { useContext } from 'react';
import { OrganizationContext } from '../contexts/organization-context';
import { BlockOrganizationSection } from './block-organization';
import { DomainsList } from './domains-list';
import { PageTitle } from '../../../../components/PageTitle';
import { useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useEffect } from 'react';

const AddDomainSection = () => {
  return (
    <Flex gap={5} justify="between">
      <Flex direction="column" gap={3}>
        <Text size="large">Allowed email domains</Text>
        <Text size="small" variant="secondary">
          Anyone with an email address at these domains is allowed to sign up
          for this workspace.
        </Text>
      </Flex>
      <Tooltip>
        <Tooltip.Trigger render={<Flex style={{ alignSelf: "center" }} />}>
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
        </Tooltip.Trigger>
        <Tooltip.Content side="bottom">Coming soon</Tooltip.Content>
      </Tooltip>
    </Flex>
  );
};

export const OrganizationSecurity = () => {
  const { organization } = useContext(OrganizationContext);

  const { data: domains, isLoading, error } = useQuery(
    FrontierServiceQueries.listOrganizationDomains,
    create(ListOrganizationDomainsRequestSchema, {
      orgId: organization?.id
    }),
    {
      enabled: !!organization?.id,
      select: data => data?.domains || []
    }
  );

  useEffect(() => {
    if (error) {
      toastManager.add({
        title: "Something went wrong",
        description: "Unable to fetch domains",
        type: "error",
      });
      console.error("Unable to fetch domains:", error);
    }
  }, [error]);

  const title = `Security | ${organization?.title} | Organizations`;

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
