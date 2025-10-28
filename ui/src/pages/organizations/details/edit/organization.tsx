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
} from "@raystack/apsara";
import { Cross1Icon } from "@radix-ui/react-icons";
import { AvatarUpload } from "@raystack/frontier/react";
import { AppContext } from "~/contexts/App";
import { z } from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import {
  type Organization,
  OrganizationSchema,
} from "@raystack/proton/frontier";
import { create, type JsonObject } from "@bufbuild/protobuf";

const orgUpdateSchema = z
  .object({
    avatar: z.string().optional(),
    title: z.string(),
    name: z.string(),
    size: z.string().transform((value) => parseInt(value)),
    type: z.string(),
    otherType: z.string().optional(),
    country: z.string(),
  })
  .refine(
    (data) =>
      data.type !== "other" || (data.type === "other" && data.otherType),
    {
      message: "otherType is required when type is 'other'.",
      path: ["otherType"],
    },
  );

type OrgUpdateSchema = z.infer<typeof orgUpdateSchema>;

async function loadCountries() {
  const data = await import("~/assets/data/countries.json");
  return data.default;
}

interface MetaData extends Record<string, unknown> {
  size?: number;
  type?: string;
  country?: string;
}

const otherTypePrefix = "Other - ";

function removeOtherPrefix(text: string) {
  if (text.startsWith(otherTypePrefix)) {
    return text.substring(otherTypePrefix.length);
  }
  return text;
}

function getDefaultValue(organization: Organization, industries: string[]) {
  const metadata = organization?.metadata as MetaData;
  const type =
    metadata?.type && industries.includes(metadata?.type)
      ? metadata?.type
      : "other";
  const otherType =
    type === "other" && metadata?.type ? removeOtherPrefix(metadata?.type) : "";
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
    watch,
    register,
    formState: { errors },
  } = useForm<OrgUpdateSchema>({
    defaultValues: organization
      ? getDefaultValue(organization, industries)
      : {},
    resolver: zodResolver(orgUpdateSchema),
  });

  const {
    mutateAsync: updateOrganizationMutation,
    error: mutationError,
    isPending: isSubmitting,
  } = useMutation(FrontierServiceQueries.updateOrganization);

  useEffect(() => {
    if (mutationError) {
      if (mutationError.message?.includes("already exists")) {
        setError("name", { message: "Organization name already exists" });
      } else {
        console.error("Unable to update organization:", mutationError);
      }
    }
  }, [mutationError, setError]);

  async function onSubmit(data: OrgUpdateSchema) {
    try {
      const payload = {
        avatar: data.avatar || "",
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

      const orgResp = await updateOrganizationMutation({
        id: orgId,
        body: payload,
      });
      const organization = orgResp.organization;
      if (organization) {
        const protoOrg = create(OrganizationSchema, {
          ...organization,
          metadata: organization.metadata as JsonObject,
        });
        updateOrganization(protoOrg);
      }
    } catch (err: unknown) {
      console.error("Unable to update organization:", err);
    }
  }

  const showOtherTypeField = watch("type", "other") === "other";

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
                        align="center"
                        gap="medium"
                        style={{ width: "100%" }}
                      >
                        <AvatarUpload {...field} data-test-id="avatar-upload" />
                        <Text>Pick a logo for your organization</Text>
                      </Flex>
                    </>
                  );
                }}
              />
              <InputField {...register("title")} label="Organization title" />
              <InputField
                {...register("name")}
                prefix={config?.app_url}
                label="Organization URL"
                helperText="This will be your organization unique web address"
                error={errors.name?.message}
              />
              <InputField
                {...register("size")}
                type="number"
                label="Organization size"
                error={errors.size?.message}
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
                      <Select
                        {...field}
                        value={field?.value?.toString()}
                        onValueChange={(value) => {
                          field?.onChange({ target: { value } });
                        }}
                      >
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
                          <Select.Item value="other">Other</Select.Item>
                        </Select.Content>
                      </Select>
                    </Flex>
                  );
                }}
              />
              {showOtherTypeField ? (
                <InputField
                  label="Organization industry (other)"
                  {...register("otherType")}
                  error={errors.otherType?.message}
                />
              ) : null}
              <Controller
                name="country"
                control={control}
                render={({ field }) => {
                  return (
                    <Flex direction="column" gap={2}>
                      <Label htmlFor="country-select">Country</Label>
                      <Select
                        {...field}
                        value={field?.value?.toString()}
                        onValueChange={(value) => {
                          field?.onChange({ target: { value } });
                        }}
                        autocomplete
                      >
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
