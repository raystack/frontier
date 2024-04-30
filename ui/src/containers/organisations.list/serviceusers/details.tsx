import { Flex, Grid, Text } from "@raystack/apsara";
import { V1Beta1Organization, V1Beta1ServiceUser } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { Link, Outlet, useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";

type DetailsProps = {
  key: string;
  value: any;
};

export default function ServiceUserDetails() {
  let { organisationId, serviceUserId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [serviceUser, setServiceUser] = useState<V1Beta1ServiceUser>();

  const { client } = useFrontier();

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
        href: `/organisations/${organisation?.id}/serviceusers/${serviceUser?.id}`,
        name: serviceUser?.title || "",
      },
    ],
  };

  useEffect(() => {
    async function getOrganization() {
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(organisationId ?? "");
      setOrganisation(organization);
    }

    async function getServiceUser(userId: string) {
      const resp = await client?.frontierServiceGetServiceUser(userId);
      const user = resp?.data?.serviceuser;
      setServiceUser(user);
    }
    if (organisationId) {
      getOrganization();
    }

    if (serviceUserId) {
      getServiceUser(serviceUserId);
    }
  }, [client, organisationId, serviceUserId]);

  const detailList: DetailsProps[] = [
    {
      key: "Title",
      value: serviceUser?.title,
    },
    {
      key: "State",
      value: serviceUser?.state,
    },
    {
      key: "Created At",
      value: (
        <Text>
          {dayjs(serviceUser?.created_at).format(DEFAULT_DATE_FORMAT)}
        </Text>
      ),
    },
  ];

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
      >
        <Link
          to={`/organisations/${organisation?.id}/serviceusers/${serviceUser?.id}/create-token`}
        >
          Generate Token
        </Link>
      </PageHeader>
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
        {detailList.map((detailItem) => (
          <Grid columns={2} gap="small" key={detailItem.key}>
            <Text size={1} weight={500}>
              {detailItem.key}
            </Text>
            <Text size={1}>{detailItem.value}</Text>
          </Grid>
        ))}
      </Flex>
      <Outlet />
    </Flex>
  );
}
