import {
  Button,
  Dialog,
  Flex,
  IconButton,
  Text,
  toast,
} from "@raystack/apsara/v1";
import { V1Beta1Domain } from "~/api/frontier";
import styles from "./security.module.css";
import { CheckCircledIcon, TrashIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { useState } from "react";
import { api } from "~/api";

interface DeleteDomainDialogProps {
  fetchDomains: () => Promise<void>;
  domain: V1Beta1Domain;
}

const DeleteDomainDialog = ({
  fetchDomains,
  domain,
}: DeleteDomainDialogProps) => {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      await api?.frontierServiceDeleteOrganizationDomain(
        domain?.org_id || "",
        domain?.id || "",
      );
      await fetchDomains();
      setIsDialogOpen(false);
    } catch (error) {
      toast.error("unable to delete domain");
    } finally {
      setIsDeleting(false);
    }
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
            loading={isDeleting}
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
  domain: V1Beta1Domain;
  fetchDomains: () => Promise<void>;
}

const DomainItem = ({ fetchDomains, domain }: DomainItemProps) => {
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
      <DeleteDomainDialog fetchDomains={fetchDomains} domain={domain} />
    </Flex>
  );
};

interface DomainsListProps {
  isLoading: boolean;
  domains: V1Beta1Domain[];
  fetchDomains: () => Promise<void>;
}

export const DomainsList = ({
  isLoading,
  domains = [],
  fetchDomains,
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
          fetchDomains={fetchDomains}
        />
      ))}
    </Flex>
  ) : null;
};
