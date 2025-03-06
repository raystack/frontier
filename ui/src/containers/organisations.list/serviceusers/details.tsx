import { Grid } from "@raystack/apsara";
import { Flex, Switch, Text, Separator } from "@raystack/apsara/v1";
import { V1Beta1ServiceUser } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import dayjs from "dayjs";
import { useContext, useEffect, useState } from "react";
import { Link, Outlet, useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";
import { AppContext } from "~/contexts/App";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";
import TokensList from "./tokens/list";
import { toast } from "sonner";

type DetailsProps = {
  key: string;
  value: any;
};

export default function ServiceUserDetails() {
  let { organisationId, serviceUserId } = useParams();
  const [serviceUser, setServiceUser] = useState<V1Beta1ServiceUser>();
  const [isSwitchActionLoading, setSwitchActionLoading] = useState(false);

  const { platformUsers, fetchPlatformUsers } = useContext(AppContext);
  const { client } = useFrontier();

  const isPlatformUser = Boolean(
    platformUsers?.serviceusers?.find(
      (user) => user?.id === serviceUserId && user?.state === "enabled"
    )
  );

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisationId}`,
        name: `Org`,
      },
      {
        href: `/organisations/${organisationId}/serviceusers`,
        name: "Service Users",
      },
      {
        href: `/organisations/${organisationId}/serviceusers/${serviceUser?.id}`,
        name: serviceUser?.title || "",
      },
    ],
  };

  useEffect(() => {
    async function getServiceUser(orgId: string, userId: string) {
      const resp = await client?.frontierServiceGetServiceUser(orgId, userId);
      const user = resp?.data?.serviceuser;
      setServiceUser(user);
    }

    if (serviceUserId && organisationId) {
      getServiceUser(organisationId, serviceUserId);
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

  const upatePlatformUser = async (value: boolean) => {
    try {
      setSwitchActionLoading(true);
      const resp = value
        ? await client?.adminServiceAddPlatformUser({
            serviceuser_id: serviceUserId,
            relation: "member",
          })
        : await client?.adminServiceRemovePlatformUser({
            serviceuser_id: serviceUserId,
          });
      if (resp?.status === 200) {
        await fetchPlatformUsers();
      }
    } catch (err: any) {
      console.error(err);
      toast.error("Something went wrong");
    } finally {
      setSwitchActionLoading(false);
    }
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
      >
        <Link
          to={`/organisations/${organisationId}/serviceusers/${serviceUser?.id}/create-token`}
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
        <Grid columns={2} gap="small">
          <Text size={1} weight={500}>
            Platform User
          </Text>
          <Switch
            // @ts-ignore
            disabled={isSwitchActionLoading}
            checked={isPlatformUser}
            onCheckedChange={upatePlatformUser}
          />
        </Grid>
      </Flex>
      <Separator />
      <TokensList
        organisationId={organisationId || ""}
        serviceUserId={serviceUser?.id || ""}
      />
      <Outlet />
    </Flex>
  );
}
