import { Cross1Icon } from "@radix-ui/react-icons";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Text,
  Select,
  Sheet,
  SidePanel,
  toast,
} from "@raystack/apsara";
import styles from "./edit.module.css";
import { useCallback, useContext, useEffect } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Skeleton from "react-loading-skeleton";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  GetBillingAccountDetailsRequestSchema,
  UpdateBillingAccountDetailsRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

interface EditBillingPanelProps {
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

export function EditBillingPanel({ onClose }: EditBillingPanelProps) {
  const { billingAccount, setBillingAccountDetails } =
    useContext(OrganizationContext);

  const organizationId = billingAccount?.org_id || "";
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

  const { data: billingDetailsData, isLoading, error, refetch: refetchBillingDetails } = useQuery(
    AdminServiceQueries.getBillingAccountDetails,
    create(GetBillingAccountDetailsRequestSchema, {
      orgId: organizationId,
      id: billingId,
    }),
    {
      enabled: !!organizationId && !!billingId,
      select: (data) => {
        const isPostpaid = data?.creditMin < 0n;
        const creditMin = isPostpaid ? data.creditMin * BigInt(-1) : data.creditMin;
        return {
          tokenPaymentType: isPostpaid ? "postpaid" as const : "prepaid" as const,
          creditMin: creditMin.toString(),
          dueInDays: data?.dueInDays.toString(),
        };
      },
    },
  );

  useEffect(() => {
    if (error) {
      toast.error("Something went wrong", {
        description: "Unable to fetch billing details",
      });
      console.error("Unable to fetch billing details:", error);
    }
  }, [error]);

  useEffect(() => {
    if (billingDetailsData) {
      reset(billingDetailsData);
    }
  }, [billingDetailsData, reset]);

  const { mutateAsync: updateBillingDetails } = useMutation(
    AdminServiceQueries.updateBillingAccountDetails,
    {
      onError: (error) => {
        toast.error("Something went wrong", {
          description: error.message,
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

      const getBillingDetailsResp = await refetchBillingDetails();
      const updatedDetails = getBillingDetailsResp?.data;
      if (updatedDetails && setBillingAccountDetails) {
        setBillingAccountDetails({
          credit_min: updatedDetails.creditMin.toString(),
          due_in_days: updatedDetails.dueInDays.toString(),
        });
      }

      toast.success("Billing details updated");
    } catch (error) {
      toast.error("Something went wrong", {
        description: "Failed to update billing details",
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
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
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
                <InputField
                  disabled={isPrepaid}
                  label="Credit limit"
                  type="text"
                  {...register("creditMin", {})}
                  error={errors?.creditMin?.message}
                />
              )}
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <InputField
                  disabled={isPrepaid}
                  label="Billing due date"
                  type="text"
                  suffix="Days"
                  {...register("dueInDays", {})}
                  error={errors?.dueInDays?.message}
                />
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
      </Sheet.Content>
    </Sheet>
  );
}
