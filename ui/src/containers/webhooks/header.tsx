import { PlusIcon } from "@radix-ui/react-icons";

import { Button, DataTable, Flex } from "@raystack/apsara";
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
        variant="secondary"
        onClick={() => navigate("/webhooks/create")}
        style={{ width: "100%" }}
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
