import { ChangeEvent, useCallback, useEffect, useMemo, useState } from "react";
import styles from "./list.module.css";
import { useTerminology } from "../../../hooks/useTerminology";
import { generateSlug, randomSuffix } from "~/admin/utils/helper";
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
import { debounce } from "lodash";
import { z } from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@connectrpc/connect-query";
import { Code } from "@connectrpc/connect";
import { AdminServiceQueries } from "@raystack/proton/frontier";

const orgCreateSchema = z.object({
  avatar: z.string().optional(),
  title: z.string().min(1, "Title is required"),
  name: z
    .string()
    .min(1, "URL is required")
    .min(3, "URL not valid, Min 3 characters allowed")
    .max(50, "URL not valid, Max 50 characters allowed")
    .regex(
      /^[a-zA-Z0-9_-]{3,50}$/,
      "Only numbers, letters, '-', and '_' are allowed. Spaces are not allowed.",
    ),
  orgOwnerEmail: z
    .string()
    .min(1, "Owner email is required")
    .email("Enter a valid email address"),
  country: z.string().min(1, "Country is required"),
});

type OrgCreateSchema = z.infer<typeof orgCreateSchema>;

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
  appUrl = "",
  countries: countriesProp = [],
  onSuccess,
}: CreateOrganizationPanelProps) {
  const t = useTerminology();
  const [countries, setCountries] = useState<string[]>(countriesProp);

  useEffect(() => {
    if (countriesProp.length > 0) {
      setCountries(countriesProp);
    }
  }, [countriesProp]);

  const {
    handleSubmit,
    control,
    setError,
    setValue,
    formState: { isSubmitting, errors, dirtyFields },
    register,
  } = useForm<OrgCreateSchema>({
    defaultValues: {
      avatar: "",
      title: "",
      name: "",
      orgOwnerEmail: "",
      country: "",
    },
    resolver: zodResolver(orgCreateSchema),
  });

  const uniqueSuffix = useMemo(() => randomSuffix(), []);

  const debouncedGenerateSlug = useMemo(
    () =>
      debounce((title: string) => {
        const baseSlug = generateSlug(title);
        setValue("name", baseSlug ? `${baseSlug}-${uniqueSuffix}` : "");
      }, 300),
    [setValue, uniqueSuffix],
  );

  useEffect(() => () => debouncedGenerateSlug.cancel(), [debouncedGenerateSlug]);

  const handleTitleChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      if (dirtyFields.name) return;
      debouncedGenerateSlug(e.target.value);
    },
    [debouncedGenerateSlug, dirtyFields.name],
  );

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
            noValidate
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
                <Input
                  {...register("title", { onChange: handleTitleChange })}
                />
              </Field>
              <Field
                label={`${t.organization({ case: "capital" })} owner email`}
                error={errors.orgOwnerEmail?.message}
                required
              >
                <Input {...register("orgOwnerEmail")} type="email"/>
              </Field>
              <Field
                label={`${t.organization({ case: "capital" })} URL`}
                description={`This will be your ${t.organization({ case: "lower" })} unique web address`}
                error={errors.name?.message}
                required
              >
                <Input {...register("name")} prefix={appUrl} />
              </Field>
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
