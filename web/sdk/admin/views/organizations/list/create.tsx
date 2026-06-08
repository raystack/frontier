import { useEffect, useState } from "react";
import styles from "./list.module.css";
import { useTerminology } from "../../../hooks/useTerminology";
import {
  Button,
  Field,
  Flex,
  IconButton,
  Input,
  Select,
  Drawer,
  SidePanel,
  Text,
} from "@raystack/apsara";
import { Cross1Icon } from "@radix-ui/react-icons";
import { ImageUpload } from "~/client/components/image-upload";
import { z } from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@connectrpc/connect-query";
import { Code } from "@connectrpc/connect";
import { AdminServiceQueries } from "@raystack/proton/frontier";

const orgCreateSchema = z
  .object({
    avatar: z.string().optional(),
    title: z.string().min(1, "Title is required"),
    name: z
      .string()
      .min(1, "URL is required")
      .regex(
        /^[a-z0-9-]+$/,
        "Use only lowercase letters, numbers, and hyphens",
      ),
    orgOwnerEmail: z
      .string()
      .min(1, "Owner email is required")
      .email("Enter a valid email address"),
    size: z
      .string()
      .min(1, "Size is required")
      .refine(
        (value) => Number.isInteger(Number(value)) && Number(value) > 0,
        "Enter a valid size greater than 0",
      )
      .transform((value) => parseInt(value)),
    type: z.string().min(1, "Industry is required"),
    otherType: z.string().optional(),
    country: z.string().min(1, "Country is required"),
  })
  .refine((data) => data.type !== "other" || Boolean(data.otherType?.trim()), {
    message: "Please specify the industry",
    path: ["otherType"],
  });

type OrgCreateSchema = z.infer<typeof orgCreateSchema>;

const otherTypePrefix = "Other - ";

export type CreateOrganizationPanelProps = {
  open?: boolean;
  onClose: () => void;
  organizationTypes?: string[];
  appUrl?: string;
  countries?: string[];
  onSuccess?: (orgId: string) => void;
};

export function CreateOrganizationPanel({
  open = false,
  onClose,
  organizationTypes = [],
  appUrl = "",
  countries: countriesProp = [],
  onSuccess,
}: CreateOrganizationPanelProps) {
  const t = useTerminology();
  const [countries, setCountries] = useState<string[]>(countriesProp);
  const industries = organizationTypes;

  useEffect(() => {
    if (countriesProp.length > 0) {
      setCountries(countriesProp);
    }
  }, [countriesProp]);

  const {
    handleSubmit,
    control,
    setError,
    formState: { isSubmitting, errors },
    watch,
    register,
  } = useForm<OrgCreateSchema>({
    defaultValues: {
      avatar: "",
      title: "",
      name: "",
      orgOwnerEmail: "",
      type: "",
      otherType: "",
      country: "",
    },
    resolver: zodResolver(orgCreateSchema),
  });

  const { mutateAsync: createOrganization, isPending } = useMutation(
    AdminServiceQueries.adminCreateOrganization,
    {
      onError: (error) => {
        if (error?.code === Code.AlreadyExists) {
          setError("name", {
            message: `${t.organization({ case: "capital" })} URL is already taken`,
          });
        } else {
          console.error("Unable to create new org:", error);
        }
      },
    },
  );

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
      if (organization?.id) {
        onSuccess?.(organization.id);
      }
    } catch (err: unknown) {
      console.error("Unable to create new org:", err);
    }
  }

  const showOtherTypeField = watch("type") === "other";

  return (
    <Drawer open={open} onOpenChange={(open) => !open && onClose()}>
      <Drawer.Content showCloseButton={false} className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-org-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title={`Add new ${t.organization({ case: "lower" })}`}
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
                    <Flex align="center" gap={5} style={{ width: "100%" }}>
                      <ImageUpload {...field} data-test-id="avatar-upload" />
                      <Text>Pick a logo for your {t.organization({ case: "lower" })}</Text>
                    </Flex>
                  );
                }}
              />
              <Field
                label={`${t.organization({ case: "capital" })} title`}
                error={errors.title?.message}
                required
              >
                <Input {...register("title")} />
              </Field>
              <Field
                label={`${t.organization({ case: "capital" })} owner`}
                error={errors.orgOwnerEmail?.message}
                required
              >
                <Input {...register("orgOwnerEmail")} type="email" />
              </Field>
              <Field
                label={`${t.organization({ case: "capital" })} URL`}
                description={`This will be your ${t.organization({ case: "lower" })} unique web address`}
                error={errors.name?.message}
                required
              >
                <Input {...register("name")} prefix={appUrl} />
              </Field>
              <Field
                label={`${t.organization({ case: "capital" })} size`}
                error={errors.size?.message}
                required
              >
                <Input {...register("size")} type="number" min={1} />
              </Field>
              <Controller
                name="type"
                control={control}
                render={({ field }) => {
                  return (
                    <Field
                      label={`${t.organization({ case: "capital" })} industry`}
                      error={errors.type?.message}
                      required
                    >
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
                    </Field>
                  );
                }}
              />
              {showOtherTypeField ? (
                <Field
                  label={`${t.organization({ case: "capital" })} industry (other)`}
                  error={errors.otherType?.message}
                >
                  <Input {...register("otherType")} />
                </Field>
              ) : null}
              <Controller
                name="country"
                control={control}
                render={({ field }) => {
                  return (
                    <Field
                      label="Country"
                      error={errors.country?.message}
                      required
                    >
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
                    </Field>
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
      </Drawer.Content>
    </Drawer>
  );
}
