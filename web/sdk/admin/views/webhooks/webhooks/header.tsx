import { PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex, DataTable } from "@raystack/apsara";
import { PageHeader } from "../../../components/PageHeader";
import styles from "./webhooks.module.css";

const pageHeader = {
  title: "Webhooks",
  breadcrumb: [] as { name: string; href?: string }[],
};

export type WebhooksHeaderProps = {
  header?: typeof pageHeader;
  onOpenCreate?: () => void;
};

export const WebhooksHeader = ({ header = pageHeader, onOpenCreate }: WebhooksHeaderProps) => {
  const handleCreate = () => onOpenCreate?.();

  return (
    <PageHeader
      title={header.title}
      breadcrumb={header.breadcrumb}
      className={styles.header}
    >
      <DataTable.Search placeholder="Search webhooks..." size="small" />
      <Button
        size="small"
        variant="text"
        color="neutral"
        leadingIcon={<PlusIcon />}
        data-test-id="admin-create-webhook-btn"
        onClick={handleCreate}
      >
        New Webhook
      </Button>
    </PageHeader>
  );
};
