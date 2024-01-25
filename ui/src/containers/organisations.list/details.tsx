import { GearIcon } from "@radix-ui/react-icons";
import { Button, Flex, Grid, Link, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";
import { Organisation } from "~/types/organisation";
import { User } from "~/types/user";

type DetailsProps = {
  key: string;
  value: any;
};

export const userColumns: ColumnDef<User, any>[] = [
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
export const projectColumns: ColumnDef<User, any>[] = [
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
  let { organisationId } = useParams();
  const { client } = useFrontier();
  const navigate = useNavigate();

  const [organisation, setOrganisation] = useState<Organisation>();
  const [orgUsers, setOrgUsers] = useState([]);
  const [orgProjects, setOrgProjects] = useState([]);
  const [orgServiceUsers, setOrgServiceUsers] = useState([]);

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

  useEffect(() => {
    async function getOrganization() {
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(organisationId ?? "");
      setOrganisation(organization);
    }
    getOrganization();
  }, [organisationId]);

  useEffect(() => {
    async function getOrganizationUser() {
      const {
        // @ts-ignore
        data: { users },
      } = await client?.frontierServiceListOrganizationUsers(
        organisationId ?? ""
      );
      setOrgUsers(users);
    }
    getOrganizationUser();
  }, [organisationId]);

  useEffect(() => {
    async function getOrganizationProjects() {
      const {
        // @ts-ignore
        data: { projects },
      } = await client?.frontierServiceListOrganizationProjects(
        organisationId ?? ""
      );
      setOrgProjects(projects);
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
      const {
        // @ts-ignore
        data: { serviceusers },
      } = await client?.frontierServiceListServiceUsers({
        org_id: organisationId ?? "",
      });
      setOrgServiceUsers(serviceusers);
    }
    getOrganizationProjects();
  }, [organisationId ?? ""]);

  const detailList: DetailsProps[] = [
    {
      key: "Name",
      value: organisation?.name,
    },
    {
      key: "Created At",
      value: new Date(organisation?.created_at as Date).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Link href={`/organisations/${organisationId}/users`}>
          {orgUsers.length}
        </Link>
      ),
    },
    {
      key: "Projects",
      value: (
        <Link href={`/organisations/${organisationId}/projects`}>
          {orgProjects.length}
        </Link>
      ),
    },
    {
      key: "Service Users",
      value: (
        <Link href={`/organisations/${organisationId}/serviceusers`}>
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
        style={{ borderBottom: "1px solid var(--border-base)" }}
      >
        <Button
          variant="secondary"
          onClick={() => unableDisableOrganization(organisation?.state)}
          style={{ width: "100%" }}
        >
          {organisation?.state === "enabled" ? "disable" : "enable"}
        </Button>
        <Button
          variant="secondary"
          onClick={() => navigate(`/organisations/${organisationId}/settings`)}
          style={{ width: "100%" }}
        >
          <Flex
            direction="column"
            align="center"
            style={{ paddingRight: "var(--pd-4)" }}
          >
            <GearIcon />
          </Flex>
          settings
        </Button>
      </PageHeader>
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
        {detailList.map((detailItem) => (
          <Grid columns={2} gap="small" key={detailItem.key}>
            <Text size={1}>{detailItem.key}</Text>
            <Text size={1}>{detailItem.value}</Text>
          </Grid>
        ))}
      </Flex>
    </Flex>
  );
}
