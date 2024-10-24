import { Flex, Grid, Table, Text } from "@raystack/apsara";
import { V1Beta1Organization, V1Beta1User } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import PageHeader from "~/components/page-header";
import styles from "./styles.module.css";

function OrganizationTable({
  organizations,
}: {
  organizations: V1Beta1Organization[];
}) {
  return (
    <Flex direction={"column"} style={{ padding: "0 24px" }}>
      <Text size={4}>Organizations</Text>
      <Table className={styles.orgsTable}>
        <Table.Header>
          <Table.Row>
            <Table.Head className={styles.tableCell}>ID</Table.Head>
            <Table.Head className={styles.tableCell}>Title</Table.Head>
            <Table.Head className={styles.tableCell}>Name</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {organizations.map((org) => {
            return (
              <Table.Row key={org?.id}>
                <Table.Cell className={styles.tableCell}>
                  <Link to={`/organisations/${org?.id}`}>{org?.id}</Link>
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {org?.title}
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {org?.name}
                </Table.Cell>
              </Table.Row>
            );
          })}
        </Table.Body>
      </Table>
    </Flex>
  );
}

export default function UserDetails() {
  const { client } = useFrontier();
  let { userId = "" } = useParams();
  const [user, setUser] = useState<V1Beta1User>();

  const [organizations, setOrganizations] = useState<V1Beta1Organization[]>([]);

  useEffect(() => {
    async function getUser() {
      const res = await client?.frontierServiceGetUser(userId);
      const user = res?.data?.user;
      setUser(user);
    }
    async function getUserOrgs() {
      const res = await client?.frontierServiceListOrganizationsByUser(userId);
      const orgs = res?.data?.organizations || [];
      setOrganizations(orgs);
    }
    getUser();
    getUserOrgs();
  }, [client, userId]);

  const pageHeader = {
    title: "Users",
    breadcrumb: [
      {
        href: `/users`,
        name: `Users list`,
      },
      {
        href: `/users/${user?.id}`,
        name: `${user?.email}`,
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
        style={{ borderBottom: "1px solid var(--border-base)" }}
      />
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
        <Grid columns={2} gap="small">
          <Text size={1}>Email</Text>
          <Text size={1}>{user?.email}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Created At</Text>
          <Text size={1}>
            {new Date(user?.created_at as any).toLocaleString("en", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </Text>
        </Grid>
      </Flex>
      <OrganizationTable organizations={organizations} />
    </Flex>
  );
}
