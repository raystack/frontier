import { Flex, DataTable, EmptyState } from "@raystack/apsara";
import { type ReactNode } from "react";
import { getColumns } from "./columns";
import styles from "./webhooks.module.css";
import { useWebhookQueries } from "./hooks/useWebhookQueries";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { PageHeader } from "../../../components/PageHeader";

export type WebhooksViewProps = {
  /** Icon rendered in the page header next to the title. */
  icon?: ReactNode;
};

export default function WebhooksView({ icon }: WebhooksViewProps = {}) {
  const {
    listWebhooks: { data: webhooksResponse, isLoading, error, isError },
  } = useWebhookQueries();

  const webhooks = webhooksResponse?.webhooks || [];

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Webhooks"
        subHeading={
          error?.message ||
          "Something went wrong while loading webhooks. Please try again."
        }
      />
    );
  }

  const columns = getColumns();

  return (
    <DataTable
      data={webhooks}
      columns={columns}
      isLoading={isLoading}
      defaultSort={{ name: "createdAt", order: "desc" }}
      mode="client"
    >
      <Flex direction="column" className={styles.tableWrapper}>
        <PageHeader
          title="Webhooks"
          icon={icon}
          breadcrumb={[]}
          className={styles.header}
        >
          <DataTable.Search placeholder="Search webhooks..." size="small" />
        </PageHeader>
        <DataTable.Content
          classNames={{ root: styles.tableRoot, table: styles.table }}
        />
      </Flex>
    </DataTable>
  );
}
