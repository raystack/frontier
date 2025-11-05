import {
  Button,
  Dialog,
  Flex,
  IconButton,
  Text,
  toast,
} from "@raystack/apsara";
import styles from "./security.module.css";
import { CheckCircledIcon, TrashIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { useState } from "react";
import { useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { FrontierServiceQueries, DeleteOrganizationDomainRequestSchema, Domain } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { OutletContext } from "../types";
import { useOutletContext } from "react-router-dom";

interface DeleteDomainDialogProps {
  domain: Domain;
}

const DeleteDomainDialog = ({
  domain,
}: DeleteDomainDialogProps) => {
  const {organization} = useOutletContext<OutletContext>();
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
        toast.success("Domain deleted");
      },
      onError: (error) => {
        toast.error("Something went wrong", {
          description: error.message,
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
    <Dialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>
        <IconButton size={3} data-test-id="delete-domain-button">
          <TrashIcon />
        </IconButton>
      </Dialog.Trigger>
      <Dialog.Content width={400}>
        <Dialog.Body>
          <Dialog.Title>Delete email domain</Dialog.Title>
          <Dialog.Description>
            Are you sure you want to delete this email domain?
          </Dialog.Description>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              color="neutral"
              variant="outline"
              data-test-id="delete-domain-cancel-button"
            >
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            color={"danger"}
            data-test-id="delete-domain-submit-button"
            loading={isPending}
            loaderText="Deleting..."
            onClick={handleDelete}
          >
            Delete
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

interface DomainItemProps {
  domain: Domain;
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
      <DeleteDomainDialog domain={domain} />
    </Flex>
  );
};

interface DomainsListProps {
  isLoading: boolean;
  domains: Domain[];
}

export const DomainsList = ({
  isLoading,
  domains = [],
}: DomainsListProps) => {
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
        <DomainItem
          key={domain?.id}
          domain={domain}
        />
      ))}
    </Flex>
  ) : null;
};
