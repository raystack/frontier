import { Cross1Icon } from "@radix-ui/react-icons";
import {
  Button,
  Field,
  Flex,
  IconButton,
  Input,
  Text,
  Select,
  Drawer,
  SidePanel,
  toastManager,
} from "@raystack/apsara-v1";
import styles from "./edit.module.css";
import { useCallback, useContext, useEffect, useMemo } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Skeleton from "react-loading-skeleton";
import { useMutation } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  UpdateBillingAccountDetailsRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

interface EditBillingPanelProps {
  open?: boolean;
  onClose: () => void;
}


const billingDetailsUpdateSchema = z
  .object({
    tokenPaymentType: z.enum(["prepaid", "postpaid"]),
    creditMin: z.string().regex(/^[0-9]+$/, "Credit limit must be a number greater than 0"),
    dueInDays: z.string().regex(/^[0-9]+$/, "Due days must be a number greater than 0"),
  })
  .refine(
    (data) => {
      if (data.tokenPaymentType === "postpaid") {
        try {
          return BigInt(data.creditMin) > 0n;
        } catch {
          return false
        }
      }
      return true;
    },
    {
      message: "Credit limit must be greater than 0 for postpaid",
      path: ["creditMin"],
    }
  ).refine(
    (data) => {
      if (data.tokenPaymentType === "postpaid") {
        try {
          return BigInt(data.dueInDays) > 0n;
        } catch {
          return false
        }
      }
      return true;
    },
    {
      message: "Due days must be greater than 0 for postpaid",
      path: ["dueInDays"],
    }
  );

type BillingDetailsForm = z.infer<typeof billingDetailsUpdateSchema>;

export function EditBillingPanel({ open = false, onClose }: EditBillingPanelProps) {
  const { 
    billingAccount,
    billingAccountDetails,
    isBillingAccountLoading: isLoading,
    fetchBillingAccountDetails,
  } = useContext(OrganizationContext);

  const organizationId = billingAccount?.orgId || "";
  const billingId = billingAccount?.id || "";

  const {
    watch,
    setValue,
    reset,
    register,
    handleSubmit,
    formState: { isSubmitting, errors },
  } = useForm<BillingDetailsForm>({
    resolver: zodResolver(billingDetailsUpdateSchema),
  });

  const billingDetailsData = useMemo(() => {
    const creditMin = billingAccountDetails?.creditMin || 0n
    const absoluteCreditMin = creditMin < 0n ? creditMin*BigInt(-1) : creditMin
    const dueInDays = billingAccountDetails?.dueInDays || 0n
    const isPostpaid = creditMin < 0n;
    return {
      tokenPaymentType: isPostpaid ? "postpaid" as const : "prepaid" as const,
      creditMin: absoluteCreditMin.toString(),
      dueInDays: dueInDays.toString(),
    };
  }, [billingAccountDetails]);

  useEffect(() => {
    if (billingDetailsData) {
      reset(billingDetailsData);
    }
  }, [billingDetailsData, reset]);

  const { mutateAsync: updateBillingDetails } = useMutation(
    AdminServiceQueries.updateBillingAccountDetails,
    {
      onError: (error) => {
        toastManager.add({
          title: "Something went wrong",
          description: error.rawMessage,
          type: "error",
        });
        console.error("Unable to update billing details:", error);
      },
    },
  );

  const onSubmit = async (data: BillingDetailsForm) => {
    try {
      // For prepaid, set values to 0; for postpaid, use form values
      const creditMinValue = data.tokenPaymentType === "prepaid" ? 0n : BigInt(data.creditMin);
      const dueInDaysValue = data.tokenPaymentType === "prepaid" ? 0n : BigInt(data.dueInDays);

      // For postpaid, creditMin should be negative for API
      const creditMinForApi = data.tokenPaymentType === "postpaid"
        ? -creditMinValue
        : creditMinValue;

      await updateBillingDetails(
        create(UpdateBillingAccountDetailsRequestSchema, {
          orgId: organizationId,
          id: billingId,
          creditMin: creditMinForApi,
          dueInDays: dueInDaysValue,
        }),
      );

      // Refetch current dialog's billing details
      fetchBillingAccountDetails()

      toastManager.add({ title: "Billing details updated", type: "success" });
    } catch (error) {
      toastManager.add({
        title: "Something went wrong",
        description: "Failed to update billing details",
        type: "error",
      });
      console.error("Failed to update billing details:", error);
    }
  };

  const onValueChange = useCallback((value: string) => {
    const paymentType = value as BillingDetailsForm["tokenPaymentType"];
    setValue("tokenPaymentType", paymentType);
    if (paymentType === "prepaid") {
      setValue("creditMin", "0");
      setValue("dueInDays", "0");
    } else {
      setValue("creditMin", "0");
      setValue("dueInDays", "30");
    }
  }, [setValue]);

  const tokenPaymentType = watch("tokenPaymentType");
  const isPrepaid = tokenPaymentType === "prepaid";

  return (
    <Drawer open={open} onOpenChange={(open) => !open && onClose()}>
      <Drawer.Content showCloseButton={false} className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-billing-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title="Edit billing"
            actions={[
              <IconButton
                key="close-billing-panel-icon"
                data-test-id="close-billing-panel-icon"
                onClick={onClose}
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />
          <form
            onSubmit={handleSubmit(onSubmit)}
            className={styles["side-panel-form"]}
          >
            <Flex
              direction="column"
              gap={5}
              className={styles["side-panel-content"]}
            >
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <Flex direction="column" gap={2}>
                  <Text variant="secondary" weight={"medium"} size="mini">
                    Token payment type
                  </Text>
                  <Select
                    value={tokenPaymentType}
                    onValueChange={onValueChange}
                  >
                    <Select.Trigger>
                      <Select.Value />
                    </Select.Trigger>
                    <Select.Content>
                      <Select.Item value="prepaid">Prepaid</Select.Item>
                      <Select.Item value="postpaid">Postpaid</Select.Item>
                    </Select.Content>
                  </Select>
                </Flex>
              )}
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <Field label="Credit limit" error={errors?.creditMin?.message}>
                  <Input
                    disabled={isPrepaid}
                    type="text"
                    {...register("creditMin", {})}
                  />
                </Field>
              )}
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <Field
                  label="Billing due date"
                  error={errors?.dueInDays?.message}
                >
                  <Input
                    disabled={isPrepaid}
                    type="text"
                    suffix="Days"
                    {...register("dueInDays", {})}
                  />
                </Field>
              )}
            </Flex>

            <Flex className={styles["side-panel-footer"]} gap={3}>
              <Button
                variant="outline"
                color="neutral"
                onClick={onClose}
                data-test-id="cancel-update-billing-details-button"
              >
                Cancel
              </Button>
              <Button
                data-test-id="save-update-billing-details-button"
                type="submit"
                disabled={isLoading || isSubmitting}
                loading={isSubmitting}
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
