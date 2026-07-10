import {
  AlertDialog,
  Button,
  Flex,
  IconButton,
  Skeleton,
  Text,
  toastManager,
} from "@raystack/apsara";
import styles from "./security.module.css";
import { CheckCircledIcon, TrashIcon } from "@radix-ui/react-icons";
import { useContext, useState } from "react";
import { useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { FrontierServiceQueries, DeleteOrganizationDomainRequestSchema, Domain } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { OrganizationContext } from "../contexts/organization-context";

interface DeleteDomainDialogProps {
  domain: Domain;
}

const DeleteDomainDialog = ({
  domain,
}: DeleteDomainDialogProps) => {
  const { organization } = useContext(OrganizationContext);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutateAsync: deleteDomain, isPending } = useMutation(
    FrontierServiceQueries.deleteOrganizationDomain,
    {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listOrganizationDomains,
            transport,
            input: { orgId: organization?.id || "" },
            cardinality: "finite",
          }),
        });
        setIsDialogOpen(false);
        toastManager.add({ title: "Domain deleted", type: "success" });
      },
      onError: (error) => {
        toastManager.add({
          title: "Something went wrong",
          description: error.rawMessage,
          type: "error",
        });
        console.error("Unable to delete domain:", error);
      },
    },
  );

  const handleDelete = async () => {
    await deleteDomain(
      create(DeleteOrganizationDomainRequestSchema, {
        orgId: domain?.orgId || "",
        id: domain?.id || "",
      }),
    );
  };

  const onOpenChange = (value: boolean) => {
    setIsDialogOpen(value);
  };

  return (
    <AlertDialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <AlertDialog.Trigger
        render={
          <IconButton size={3} data-test-id="delete-domain-button">
            <TrashIcon />
          </IconButton>
        }
      />
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>Delete email domain</AlertDialog.Title>
          <AlertDialog.Description>
            Are you sure you want to delete this email domain?
          </AlertDialog.Description>
        </AlertDialog.Header>
        <AlertDialog.Footer>
          <AlertDialog.Close
            render={
              <Button
                color="neutral"
                variant="outline"
                data-test-id="delete-domain-cancel-button"
              >
                Cancel
              </Button>
            }
          />
          <Button
            color={"danger"}
            data-test-id="delete-domain-submit-button"
            loading={isPending}
            loaderText="Deleting..."
            onClick={handleDelete}
          >
            Delete
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};

interface DomainItemProps {
  domain: Domain;
}

const DomainItem = ({ domain }: DomainItemProps) => {
  return (
    <Flex className={styles["domains-list-item"]} justify="between">
      <Flex gap={3}>
        <Text size="small">{domain?.name}</Text>
        {domain.state === "verified" ? (
          <CheckCircledIcon
            color={"var(--rs-color-foreground-success-primary)"}
          />
        ) : null}
      </Flex>
      <DeleteDomainDialog domain={domain} />
    </Flex>
  );
};

interface DomainsListProps {
  isLoading: boolean;
  domains: Domain[];
  error?: Error | null;
  onRetry: () => void;
}

export const DomainsList = ({
  isLoading,
  domains = [],
  error,
  onRetry,
}: DomainsListProps) => {
  if (!isLoading && error) {
    return (
      <Flex
        justify="between"
        align="center"
        className={`${styles["domains-list"]} ${styles["domains-list-status"]}`}
      >
        <Text size="small" variant="secondary">
          Couldn&apos;t load email domains
        </Text>
        <Button
          variant="outline"
          color="neutral"
          size="small"
          onClick={onRetry}
          data-test-id="retry-domains-button"
        >
          Retry
        </Button>
      </Flex>
    );
  }

  return isLoading ? (
    <Flex
      justify="between"
      align="center"
      className={`${styles["domains-list"]} ${styles["domains-list-status"]}`}
    >
      <Skeleton height={16} width={180} />
      <Skeleton height={20} width={20} borderRadius="var(--rs-radius-2)" />
    </Flex>
  ) : domains.length ? (
    <Flex direction="column" className={styles["domains-list"]}>
      {domains.map((domain) => (
        <DomainItem
          key={domain?.id}
          domain={domain}
        />
      ))}
    </Flex>
  ) : (
    <Flex className={`${styles["domains-list"]} ${styles["domains-list-empty"]}`}>
      <Text size="small" variant="secondary">
        No allowed email domains added yet
      </Text>
    </Flex>
  );
};
