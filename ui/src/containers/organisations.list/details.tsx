import {
  Button,
  Dialog,
  Flex,
  Grid,
  Select,
  Separator,
  Switch,
  Text,
} from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import { User } from "~/types/user";
import { V1Beta1BillingAccount, V1Beta1Organization } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";
import { useParams } from "react-router-dom";
import { Cross1Icon } from "@radix-ui/react-icons";
import * as zod from "zod";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as Form from "@radix-ui/react-form";
import { CustomFieldName } from "~/components/CustomField";

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

const billingAccountSchema = zod
  .object({
    name: zod.string(),
    email: zod.string().trim().email(),
    address: zod
      .object({
        country: zod.string(),
      })
      .required(),
    currency: zod.string(),
  })
  .required();

type BillingAccountFormSchema = zod.infer<typeof billingAccountSchema>;

const billingFormfields = [
  {
    key: "name",
  },
  {
    key: "email",
  },
  {
    key: "address.country",
    label: "Country",
  },
  {
    key: "currency",
    defaultValue: "USD",
  },
];

interface BillingAccountFormProps {
  organization?: V1Beta1Organization;
}

function BillingAccountForm({ organization }: BillingAccountFormProps) {
  const methods = useForm<BillingAccountFormSchema>({
    resolver: zodResolver(billingAccountSchema),
    defaultValues: { name: organization?.title },
  });

  function onSubmit(data: BillingAccountFormSchema) {}

  function onDialogOpen() {
    methods.reset();
  }

  console.log(methods);
  return (
    <Dialog onOpenChange={onDialogOpen}>
      <Dialog.Trigger asChild>
        <Button>+</Button>
      </Dialog.Trigger>
      <Dialog.Content>
        <Flex direction={"column"} gap={"small"}>
          <Flex align={"center"} justify={"between"}>
            <Text size={5} style={{ fontWeight: 500 }}>
              Create Billing Account
            </Text>
            <Dialog.Close className={"closeBtn"}>
              <Cross1Icon />
            </Dialog.Close>
          </Flex>
          <Separator />
          <FormProvider {...methods}>
            <Form.Root onSubmit={methods.handleSubmit(onSubmit)}>
              <Flex direction={"column"} gap="medium">
                {billingFormfields.map((field) => {
                  return (
                    <CustomFieldName
                      key={field.key}
                      label={field.label}
                      name={field.key}
                      register={methods.register}
                      control={methods.control}
                      defaultValue={field.defaultValue}
                    />
                  );
                })}
                <Separator />
                <Flex gap="small" justify={"end"}>
                  <Form.Submit asChild>
                    <Button variant={"primary"}>Create</Button>
                  </Form.Submit>
                  <Dialog.Close asChild>
                    <Button type="reset">Cancel</Button>
                  </Dialog.Close>
                </Flex>
              </Flex>
            </Form.Root>
          </FormProvider>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}

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
        const org = resp?.data?.organization;
        setOrganisation(org);
        getOrganizationProjects(org?.id || "");
        getOrganizationUser(org?.id || "");
        getOrganizationBillingAccounts(org?.id || "");
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsOrganisationLoading(false);
    }
  }

  async function getOrganizationUser(orgId: string) {
    const {
      // @ts-ignore
      data: { users },
    } = await client?.frontierServiceListOrganizationUsers(orgId);
    setOrgUsers(users);
  }

  async function getOrganizationProjects(orgId: string) {
    const {
      // @ts-ignore
      data: { projects },
    } = await client?.frontierServiceListOrganizationProjects(orgId);
    setOrgProjects(projects);
  }

  async function getOrganizationBillingAccounts(orgId: string) {
    setIsBillingAccountsLoading(true);
    try {
      const resp = await client?.frontierServiceListBillingAccounts(orgId);
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

  async function createBillingAccount(params: any) {
    await client?.frontierServiceCreateBillingAccount(
      organisation?.id || "",
      {}
    );
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
            <BillingAccountForm organization={organisation} />
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
