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
} from "@raystack/apsara";
import styles from "./edit.module.css";
import { z } from "zod";
import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AdminServiceQueries, FrontierServiceQueries, SetOrganizationKycRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

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
  const { organization, kycDetails, updateKYCDetails } = useContext(OrganizationContext);
  const queryClient = useQueryClient();
  const transport = useTransport();
  const orgId = organization?.id || "";

  const {
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<KYCUpdateSchema>({
    defaultValues: {
      status: kycDetails?.status || false,
      link: kycDetails?.link || "",
    },
    resolver: zodResolver(kycUpdateSchema),
  });

  const {
    mutateAsync: setOrganizationKycMutation,
  } = useMutation(AdminServiceQueries.setOrganizationKyc, {
    onSuccess: (data) => {
      // Manually update context state for immediate UI update
      // TODO: Remove this once OrganizationContext uses useQuery for KYC data
      // and rely solely on query invalidation below
      const newKycDetails = data.organizationKyc;
      if (newKycDetails) {
        updateKYCDetails(newKycDetails);
      }
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.getOrganizationKyc,
          transport,
          input: {orgId},
          cardinality: "finite",
        })
      });
      toast.success("KYC details updated successfully");
      onClose();
    },
    onError: (error) => {
      toast.error(`Failed to update KYC details: ${error.message}`);
      console.error("Unable to update KYC details:", error);
    },
  });

  async function submit(data: KYCUpdateSchema) {
    if (!orgId) {
      return;
    }
    await setOrganizationKycMutation(
      create(SetOrganizationKycRequestSchema, {
        orgId: orgId,
        status: data.status,
        link: data.link || "",
      })
    );
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
