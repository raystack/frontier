import {
  Button,
  Dialog,
  Flex,
  Grid,
  Select,
  Switch,
  Text,
} from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import { User } from "~/types/user";
import { useOrganisation } from ".";
import { V1Beta1BillingAccount, V1Beta1Organization } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";
import { useParams } from "react-router-dom";

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
  const { client } = useFrontier();
  const { organisationId } = useParams();

  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [isOrganisationLoading, setIsOrganisationLoading] = useState(false);

  const [orgUsers, setOrgUsers] = useState([]);
  const [orgProjects, setOrgProjects] = useState([]);
  const [billingAccounts, setBillingAccounts] = useState<
    V1Beta1BillingAccount[]
  >([]);
  const [isBillingAccountsLoading, setIsBillingAccountsLoading] =
    useState(false);

  async function getOrganization() {
    setIsOrganisationLoading(true);
    try {
      const resp = await client?.frontierServiceGetOrganization(
        organisationId || ""
      );
      if (resp?.data?.organization) {
        setOrganisation(resp?.data?.organization);
        getOrganizationProjects();
        getOrganizationUser();
        getOrganizationBillingAccounts();
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsOrganisationLoading(false);
    }
  }

  async function getOrganizationUser() {
    const {
      // @ts-ignore
      data: { users },
    } = await client?.frontierServiceListOrganizationUsers(
      organisation?.id ?? ""
    );
    setOrgUsers(users);
  }

  async function getOrganizationProjects() {
    const {
      // @ts-ignore
      data: { projects },
    } = await client?.frontierServiceListOrganizationProjects(
      organisation?.id ?? ""
    );
    setOrgProjects(projects);
  }

  async function getOrganizationBillingAccounts() {
    setIsBillingAccountsLoading(true);
    try {
      const resp = await client?.frontierServiceListBillingAccounts(
        organisation?.id ?? ""
      );
      if (resp?.data?.billing_accounts) {
        setBillingAccounts(resp?.data?.billing_accounts);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsBillingAccountsLoading(false);
    }
  }

  useEffect(() => {
    if (organisationId) {
      getOrganization();
    }
  }, [organisationId]);

  const detailList: DetailsProps[] = [
    {
      key: "Id",
      value: organisation?.id,
    },
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
      value: new Date(organisation?.created_at || "").toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Dialog>
          <Dialog.Trigger asChild>
            <Button>{orgUsers.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={userColumns}
              data={orgUsers}
              header={<DialogHeader title="Organization users" />}
            />
          </Dialog.Content>
        </Dialog>
      ),
    },
    {
      key: "Projects",
      value: (
        <Dialog>
          <Dialog.Trigger asChild>
            <Button>{orgProjects.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={projectColumns}
              data={orgProjects}
              header={<DialogHeader title="Organization project" />}
            />
          </Dialog.Content>
        </Dialog>
      ),
    },
  ];

  async function onOrgStateChange(value: boolean) {
    setIsOrganisationLoading(true);
    try {
      const resp = value
        ? await client?.frontierServiceEnableOrganization(
            organisation?.id || "",
            {}
          )
        : await client?.frontierServiceDisableOrganization(
            organisation?.id || "",
            {}
          );
      if (resp?.data) {
        getOrganization();
      }
    } catch (err) {
      console.error(err);
    }
  }

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "320px",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
        padding: "var(--pd-16)",
      }}
    >
      <Text size={4}>{organisation?.name}</Text>
      <Flex direction="column" gap="large">
        {detailList.map((detailItem) => (
          <Grid columns={2} gap="small" key={detailItem.key}>
            <Text size={1} style={{ fontWeight: 500 }}>
              {detailItem.key}
            </Text>
            {isOrganisationLoading ? (
              <Skeleton />
            ) : (
              <Text size={1}>{detailItem.value}</Text>
            )}
          </Grid>
        ))}
        <Flex direction={"column"} gap="small">
          <Text size={2} style={{ fontWeight: 500 }}>
            State
          </Text>
          {isOrganisationLoading ? (
            <Skeleton />
          ) : (
            <Flex align={"center"} gap="medium">
              <Text>Disabled</Text>
              <Switch
                // @ts-ignore
                checked={organisation?.state === "enabled"}
                onCheckedChange={onOrgStateChange}
              />
              <Text>Enabled</Text>
            </Flex>
          )}
        </Flex>
        <Flex direction={"column"} gap="small">
          <Flex justify={"between"} align={"center"}>
            <Text size={2} style={{ fontWeight: 500 }}>
              Billing Accounts
            </Text>
            <Button variant={"primary"}>+</Button>
          </Flex>
          {isBillingAccountsLoading || isOrganisationLoading ? (
            <Skeleton />
          ) : (
            <Select>
              <Select.Trigger
                style={{ minWidth: "120px" }}
                disabled={billingAccounts.length === 0}
              >
                <Select.Value placeholder="Select Billing account" />
              </Select.Trigger>
              <Select.Content>
                {billingAccounts.map((billingAccount) => (
                  <Select.Item
                    key={billingAccount?.id}
                    value={billingAccount?.id}
                  >
                    {billingAccount?.name || billingAccount?.id}
                  </Select.Item>
                ))}
              </Select.Content>
            </Select>
          )}
        </Flex>
      </Flex>
    </Flex>
  );
}
