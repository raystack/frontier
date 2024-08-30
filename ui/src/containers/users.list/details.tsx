import { Flex, Grid, Text } from "@raystack/apsara";
import { V1Beta1User } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";

export default function UserDetails() {
  const { client } = useFrontier();
  let { userId } = useParams();
  const [user, setUser] = useState<V1Beta1User>();

  useEffect(() => {
    async function getProject() {
      const {
        // @ts-ignore
        data: { user },
      } = await client?.frontierServiceGetUser(userId ?? "") || {};
      setUser(user);
    }
    getProject();
  }, [userId]);

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
    </Flex>
  );
}
