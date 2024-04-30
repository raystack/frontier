import { Flex } from "@raystack/apsara";
import { V1Beta1Organization } from "@raystack/frontier";
import { useState } from "react";
import { useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";

export default function ServiceUserDetails() {
  let { organisationId, serviceUserID } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisation?.id}`,
        name: `${organisation?.title}`,
      },
      {
        href: `/organisations/${organisation?.id}/serviceusers`,
        name: "Service Users",
      },
      {
        href: `/organisations/${organisation?.id}/serviceusers`,
        name: "Service Users",
      },
    ],
  };

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "100%",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
      }}
    >
      <PageHeader
        title={pageHeader.title}
        breadcrumb={pageHeader.breadcrumb}
        style={{ borderBottom: "1px solid var(--border-base)", gap: "16px" }}
      ></PageHeader>
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}></Flex>
    </Flex>
  );
}
