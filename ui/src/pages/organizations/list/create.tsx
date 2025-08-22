import { useContext, useEffect, useState } from "react";
import styles from "./list.module.css";
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
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useNavigate } from "react-router-dom";

const orgCreateSchema = z
  .object({
    avatar: z.string().optional(),
    title: z.string(),
    name: z.string(),
    orgOwnerEmail: z.string().email(),
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

type OrgCreateSchema = z.infer<typeof orgCreateSchema>;

async function loadCountries() {
  const data = await import("~/assets/data/countries.json");
  return data.default;
}

const otherTypePrefix = "Other - ";

export function CreateOrganizationPanel({ onClose }: { onClose: () => void }) {
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);
  const navigate = useNavigate();

  const industries = config?.organization_types || [];

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
    watch,
    register,
  } = useForm<OrgCreateSchema>({
    defaultValues: {},
    resolver: zodResolver(orgCreateSchema),
  });

  const {
    mutateAsync: createOrganization,
    error: mutationError,
    isPending,
  } = useMutation(AdminServiceQueries.adminCreateOrganization);

  useEffect(() => {
    if (mutationError) {
      if (mutationError.message?.includes("already exists")) {
        setError("name", { message: "Organization name already exists" });
      } else {
        console.error("Unable to create new org:", mutationError);
      }
    }
  }, [mutationError, setError]);

  async function onSubmit(data: OrgCreateSchema) {
    try {
      const payload = {
        avatar: data.avatar || "",
        name: data.name,
        title: data.title,
        orgOwnerEmail: data.orgOwnerEmail,
        metadata: {
          size: data.size.toString(),
          type: data.otherType
            ? `${otherTypePrefix}${data.otherType}`
            : data.type,
          country: data.country,
        },
      };

      const orgResp = await createOrganization({ body: payload });
      const organization = orgResp.organization;
      if (organization) {
        navigate(`/organizations/${organization.id}`);
      }
    } catch (err: unknown) {
      console.error("Unable to create new org:", err);
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
            title="Add new organization"
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
              gap={8}
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
                {...register("orgOwnerEmail")}
                label="Organization owner"
                error={errors.orgOwnerEmail?.message}
              />
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
                loading={isSubmitting || isPending}
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
