import { useContext, useEffect, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import styles from "./edit.module.css";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Select,
  Sheet,
  SidePanel,
  Text,
  Label,
} from "@raystack/apsara/v1";
import { Cross1Icon } from "@radix-ui/react-icons";
import { AvatarUpload } from "@raystack/frontier/react";
import { AppContext } from "~/contexts/App";
import { z } from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { api } from "~/api";
import {
  Frontierv1Beta1OrganizationRequestBody,
  V1Beta1Organization,
} from "~/api/frontier";

const orgUpdateSchema = z.object({
  avatar: z.string().optional(),
  title: z.string(),
  name: z.string(),
  size: z.string().transform((value) => parseInt(value)),
  type: z.string(),
  otherType: z.string(),
  country: z.string(),
});

type OrgUpdateSchema = z.infer<typeof orgUpdateSchema>;

async function loadCountries() {
  const data = await import("~/assets/data/countries.json");
  return data.default;
}

interface MetaData {
  size: number;
  type: string;
  country: string;
}

const otherTypePrefix = "Other - ";

function removeOtherPrefix(text: string) {
  if (text.startsWith(otherTypePrefix)) {
    return text.substring(otherTypePrefix.length);
  }
  return text;
}

function getDefaultValue(
  organization: V1Beta1Organization,
  industries: string[],
) {
  const metadata = organization?.metadata as MetaData;
  const type =
    metadata?.type && industries.includes(metadata?.type)
      ? metadata?.type
      : "other";
  const otherType = type === "other" ? removeOtherPrefix(metadata?.type) : "";
  return {
    title: organization?.title || "",
    name: organization?.name || "",
    size: metadata?.size || 0,
    type: type,
    otherType: otherType,
    country: metadata?.country || "",
  };
}

export function EditOrganizationPanel({ onClose }: { onClose: () => void }) {
  const { config } = useContext(AppContext);
  const { organization, updateOrganization } = useContext(OrganizationContext);
  const [countries, setCountries] = useState<string[]>([]);

  const industries = config?.organization_types || [];
  const orgId = organization?.id || "";

  useEffect(() => {
    loadCountries().then((list) =>
      setCountries(list.map((country) => country.name)),
    );
  }, []);

  const {
    handleSubmit,
    control,
    setError,
    formState: { isSubmitting, errors },
  } = useForm<OrgUpdateSchema>({
    defaultValues: organization
      ? getDefaultValue(organization, industries)
      : {},
    resolver: zodResolver(orgUpdateSchema),
  });

  async function onSubmit(data: OrgUpdateSchema) {
    try {
      const payload: Frontierv1Beta1OrganizationRequestBody = {
        avatar: data.avatar,
        name: data.name,
        title: data.title,
        metadata: {
          size: data.size.toString(),
          type: data.otherType
            ? `${otherTypePrefix}${data.otherType}`
            : data.type,
          country: data.country,
        },
      };

      const orgResp = await api.frontierServiceUpdateOrganization(
        orgId,
        payload,
      );
      const organization = orgResp?.data?.organization;
      if (organization) {
        updateOrganization(organization);
      }
    } catch (err: unknown) {
      if (err instanceof Response && err?.status === 409) {
        setError("name", { message: "Organization name already exists" });
      } else {
        console.error(err);
      }
    }
  }

  return (
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-org-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title="Edit organization"
            actions={[
              <IconButton
                key="close-edit-org-panel-icon"
                data-test-id="close-edit-org-panel-icon"
                onClick={onClose}
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />
          <form
            className={styles["side-panel-form"]}
            onSubmit={handleSubmit(onSubmit)}
          >
            <Flex
              direction="column"
              gap={9}
              className={styles["side-panel-content"]}
            >
              <Controller
                name="avatar"
                control={control}
                render={({ field }) => {
                  return (
                    <>
                      <Flex
                        align={"center"}
                        gap={"medium"}
                        style={{ width: "100%" }}
                      >
                        <AvatarUpload {...field} data-test-id="avatar-upload" />
                        <Text>Pick a logo for your organization</Text>
                      </Flex>
                    </>
                  );
                }}
              />
              <Controller
                name="title"
                control={control}
                render={({ field }) => {
                  return <InputField {...field} label="Organization title" />;
                }}
              />

              <Controller
                name="name"
                control={control}
                render={({ field }) => {
                  return (
                    <InputField
                      {...field}
                      prefix={config?.app_url}
                      label="Organization URL"
                      helperText="This will be your organization unique web address"
                      error={errors.name?.message}
                    />
                  );
                }}
              />

              <Controller
                name="size"
                control={control}
                render={({ field }) => {
                  return (
                    <InputField
                      {...field}
                      type="number"
                      label="Organization size"
                      error={errors.size?.message}
                    />
                  );
                }}
              />

              <Controller
                name="type"
                control={control}
                render={({ field }) => {
                  return (
                    <Flex direction="column" gap={2}>
                      <Label htmlFor="org-type-select">
                        Organization industry
                      </Label>
                      <Select {...field} value={field?.value?.toString()}>
                        <Select.Trigger>
                          <Select.Value
                            placeholder="Select an industry"
                            className={styles["select-value"]}
                            id="org-type-select"
                          />
                        </Select.Trigger>
                        <Select.Content className={styles["select-content"]}>
                          {industries.map((industry) => (
                            <Select.Item key={industry} value={industry}>
                              {industry}
                            </Select.Item>
                          ))}
                          <Select.Item value={"other"}>Other</Select.Item>
                        </Select.Content>
                      </Select>
                    </Flex>
                  );
                }}
              />
              <Controller
                name="otherType"
                control={control}
                render={({ field }) => (
                  <InputField
                    label="Organization industry (other)"
                    {...field}
                    error={errors.otherType?.message}
                  />
                )}
              />
              <Controller
                name="country"
                control={control}
                render={({ field }) => {
                  return (
                    <Flex direction="column" gap={2}>
                      <Label htmlFor="country-select">Country</Label>
                      <Select {...field} value={field?.value?.toString()}>
                        <Select.Trigger>
                          <Select.Value
                            placeholder="Select a country"
                            className={styles["select-value"]}
                            id="country-select"
                          />
                        </Select.Trigger>
                        <Select.Content className={styles["select-content"]}>
                          {countries.map((country) => (
                            <Select.Item key={country} value={country}>
                              {country}
                            </Select.Item>
                          ))}
                        </Select.Content>
                      </Select>
                    </Flex>
                  );
                }}
              />
            </Flex>
            <Flex className={styles["side-panel-footer"]} gap={3}>
              <Button
                variant="outline"
                color="neutral"
                onClick={onClose}
                data-test-id="cancel-edit-org-button"
              >
                Cancel
              </Button>
              <Button
                loading={isSubmitting}
                data-test-id="save-edit-org-button"
                type="submit"
                loaderText="Saving..."
              >
                Save
              </Button>
            </Flex>
          </form>
        </SidePanel>
      </Sheet.Content>
    </Sheet>
  );
}
