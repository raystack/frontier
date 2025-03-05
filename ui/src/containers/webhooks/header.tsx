import { PlusIcon } from "@radix-ui/react-icons";

import { DataTable } from "@raystack/apsara";
import { Button, Flex } from "@raystack/apsara/v1";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const pageHeader = {
  title: "Webhooks",
  breadcrumb: [],
};
export const WebhooksHeader = ({ header = pageHeader }: any) => {
  const navigate = useNavigate();

  return (
    <PageHeader title={header.title} breadcrumb={header.breadcrumb}>
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search webhooks..." />
      <Button
        size={"small"}
        color="neutral"
        variant={"outline"}
        onClick={() => navigate("/webhooks/create")}
        style={{ width: "100%" }}
        data-test-id="admin-ui-create-webhook-btn"
      >
        <Flex
          direction="column"
          align="center"
          style={{ paddingRight: "var(--pd-4)" }}
        >
          <PlusIcon />
        </Flex>
        New webhook
      </Button>
    </PageHeader>
  );
};
