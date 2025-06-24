import { PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex, DataTable } from "@raystack/apsara/v1";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";
import styles from "./webhooks.module.css";

const pageHeader = {
  title: "Webhooks",
  breadcrumb: [],
};
export const WebhooksHeader = ({ header = pageHeader }: any) => {
  const navigate = useNavigate();

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
        data-test-id="admin-ui-create-webhook-btn"
        onClick={() => navigate("/webhooks/create")}
      >
        New Webhook
      </Button>
    </PageHeader>
  );
};
