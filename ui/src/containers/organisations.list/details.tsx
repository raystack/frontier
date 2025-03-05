import { Grid } from "@raystack/apsara";
import { Button, Flex, Separator, Text } from "@raystack/apsara/v1";
import {
  V1Beta1Organization,
  V1Beta1Project,
  V1Beta1ServiceUser,
  V1Beta1User,
} from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, NavLink, useNavigate, useParams } from "react-router-dom";

import PageHeader from "~/components/page-header";
import { capitalizeFirstLetter } from "~/utils/helper";

type DetailsProps = {
  key: string;
  value: any;
};

export const userColumns: ColumnDef<V1Beta1User, any>[] = [
  {
    header: "Name",
    accessorKey: "name",
    cell: (info) => info.getValue(),
  },
  {
    header: "Email",
    accessorKey: "email",
    cell: (info) => info.getValue(),
  },
];
export const projectColumns: ColumnDef<V1Beta1User, any>[] = [
  {
    header: "Name",
    accessorKey: "name",
    cell: (info) => info.getValue(),
  },
  {
    header: "Slug",
    accessorKey: "slug",
    cell: (info) => info.getValue(),
  },
];

export default function OrganisationDetails() {
  let organisationId = useParams()?.organisationId ?? "";
  const { client } = useFrontier();
  const navigate = useNavigate();

  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [orgUsers, setOrgUsers] = useState<V1Beta1User[]>([]);
  const [orgProjects, setOrgProjects] = useState<V1Beta1Project[]>([]);
  const [orgServiceUsers, setOrgServiceUsers] = useState<V1Beta1ServiceUser[]>(
    []
  );

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisation?.id}`,
        name: `${organisation?.name}`,
      },
    ],
  };

  async function getOrganization() {
    try {
      const res = await client?.frontierServiceGetOrganization(organisationId);
      const organization = res?.data?.organization;

      setOrganisation(organization);
    } catch (error) {
      console.error(error);
    }
  }

  async function getOrganizationUser() {
    try {
      const res = await client?.frontierServiceListOrganizationUsers(
        organisationId
      );
      const users = res?.data.users ?? [];
      setOrgUsers(users);
    } catch (error) {
      console.error(error);
    }
  }

  useEffect(() => {
    getOrganization();
  }, [organisationId]);

  useEffect(() => {
    getOrganizationUser();
  }, [organisationId]);

  useEffect(() => {
    async function getOrganizationProjects() {
      try {
        const res = await client?.frontierServiceListOrganizationProjects(
          organisationId
        );
        const projects = res?.data.projects ?? [];
        setOrgProjects(projects);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganizationProjects();
  }, [organisationId ?? ""]);

  const unableDisableOrganization = useCallback(
    async (state: string = "") => {
      if (organisationId) {
        if (state == "enabled") {
          await client?.frontierServiceDisableOrganization(organisationId, {});
        } else {
          await client?.frontierServiceEnableOrganization(organisationId, {});
        }
        navigate(0);
      }
    },
    [organisationId]
  );

  useEffect(() => {
    async function getOrganizationProjects() {
      try {
        const res = await client?.frontierServiceListServiceUsers({
          org_id: organisationId,
        });
        const serviceUsers = res?.data?.serviceusers ?? [];
        setOrgServiceUsers(serviceUsers);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganizationProjects();
  }, [organisationId ?? ""]);

  const metadataList = useMemo(() => {
    const metadata = (organisation?.metadata as Record<string, string>) || {};
    return Object.entries(metadata).map(([key, value]) => {
      return {
        key: capitalizeFirstLetter(key),
        value,
      };
    });
  }, [organisation?.metadata]);

  const detailList: DetailsProps[] = [
    {
      key: "Title",
      value: organisation?.title,
    },
    {
      key: "Name",
      value: organisation?.name,
    },
    {
      key: "Created At",
      value: new Date(organisation?.created_at as any).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Link to={`/organisations/${organisationId}/users`}>
          {orgUsers.length}
        </Link>
      ),
    },
    {
      key: "Projects",
      value: (
        <Link to={`/organisations/${organisationId}/projects`}>
          {orgProjects.length}
        </Link>
      ),
    },
    {
      key: "Service Users",
      value: (
        <Link to={`/organisations/${organisationId}/serviceusers`}>
          {orgServiceUsers.length}
        </Link>
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
        <Flex gap="medium">
          <NavLink
            to={`/organisations/${organisationId}/users`}
            style={{
              display: "flex",
              alignItems: "center",
              flexDirection: "row",
            }}
          >
            Users
          </NavLink>
          <NavLink
            to={`/organisations/${organisationId}/projects`}
            style={{
              display: "flex",
              alignItems: "center",
              flexDirection: "row",
            }}
          >
            Projects
          </NavLink>
          <NavLink
            to={`/organisations/${organisationId}/serviceusers`}
            style={{
              display: "flex",
              alignItems: "center",
              flexDirection: "row",
            }}
          >
            <span style={{ width: "max-content" }}>Service Users</span>
          </NavLink>
          <NavLink
            to={`/organisations/${organisationId}/billingaccounts`}
            style={{
              display: "flex",
              alignItems: "center",
              flexDirection: "row",
            }}
          >
            <span style={{ width: "max-content" }}>Billing Accounts</span>
          </NavLink>

          <Button
            variant="outline"
            color="neutral"
            size={"small"}
            onClick={() => unableDisableOrganization(organisation?.state)}
            style={{ width: "100%" }}
            data-test-id="admin-ui-enable-disable-org-btn"
          >
            {organisation?.state === "enabled" ? "disable" : "enable"}
          </Button>
        </Flex>
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
        {metadataList?.length > 0 ? (
          <>
            <Separator />
            <Text size={2} weight={500}>
              Metadata
            </Text>
            {metadataList.map((detailItem) => (
              <Grid columns={2} gap="small" key={detailItem.key}>
                <Text size={1} weight={500}>
                  {detailItem.key}
                </Text>
                <Text size={1}>{detailItem.value}</Text>
              </Grid>
            ))}
          </>
        ) : null}
      </Flex>
    </Flex>
  );
}
