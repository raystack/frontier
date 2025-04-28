import { Cross1Icon } from "@radix-ui/react-icons";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Label,
  Sheet,
  SidePanel,
  Switch,
  Text,
  toast,
} from "@raystack/apsara/v1";
import styles from "./edit.module.css";
import { z } from "zod";
import { api } from "~/api";
import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AxiosError } from "axios";

interface EditKYCPanelProps {
  onClose: () => void;
}

const kycUpdateSchema = z
  .object({
    status: z.boolean(),
    link: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.status && !data.link) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Link is required when status is true",
        path: ["link"],
      });
    }
  });

type KYCUpdateSchema = z.infer<typeof kycUpdateSchema>;

export function EditKYCPanel({ onClose }: EditKYCPanelProps) {
  const { organization, kycDetails, updateKYCDetails } =
    useContext(OrganizationContext);

  const {
    handleSubmit,
    control,
    formState: { isSubmitting, errors },
  } = useForm<KYCUpdateSchema>({
    defaultValues: {
      status: kycDetails?.status || false,
      link: kycDetails?.link || "",
    },
    resolver: zodResolver(kycUpdateSchema),
  });

  async function submit(data: KYCUpdateSchema) {
    if (!organization?.id) {
      return;
    }
    try {
      const result = await api?.adminServiceSetOrganizationKyc(
        organization.id,
        data,
      );
      const newKycDetails = result?.data?.organization_kyc;
      if (newKycDetails) {
        updateKYCDetails(newKycDetails);
      }
      toast.success("KYC details updated successfully");
    } catch (error) {
      const resp = (error as AxiosError<{ message: string }>)?.response?.data
        ?.message;
      toast.error(`Failed to update KYC details: ${resp}`);
      console.error(error);
    }
  }

  return (
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-kyc-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title="Edit KYC"
            actions={[
              <IconButton
                key="close-kyc-panel-icon"
                data-test-id="close-kyc-panel-icon"
                onClick={onClose}
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />

          <form
            onSubmit={handleSubmit(submit)}
            className={styles["side-panel-form"]}
          >
            <Flex
              direction="column"
              gap={5}
              className={styles["side-panel-content"]}
            >
              <Text size="small" weight="medium">
                KYC Details
              </Text>
              <Controller
                name="status"
                control={control}
                render={({ field }) => {
                  return (
                    <>
                      <Flex justify="between">
                        <Label>Mark KYC as verified</Label>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </Flex>
                      {errors.status && (
                        <Text variant="danger">{errors.status.message}</Text>
                      )}
                    </>
                  );
                }}
              />
              <Controller
                name="link"
                control={control}
                render={({ field }) => {
                  return (
                    <>
                      <InputField
                        label="Document Link"
                        {...field}
                        error={errors.link?.message}
                      />
                    </>
                  );
                }}
              />
            </Flex>
            <Flex className={styles["side-panel-footer"]} gap={3}>
              <Button
                variant="outline"
                color="neutral"
                onClick={onClose}
                data-test-id="cancel-kyc-button"
              >
                Cancel
              </Button>
              <Button
                data-test-id="save-kyc-button"
                type="submit"
                loading={isSubmitting}
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
