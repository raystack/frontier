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
import { useContext, useEffect } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Skeleton from "react-loading-skeleton";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  GetBillingAccountDetailsRequestSchema,
  GetBillingAccountRequestSchema,
  UpdateBillingAccountDetailsRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

interface EditBillingPanelProps {
  onClose: () => void;
}


const billingDetailsUpdateSchema = z
  .object({
    tokenPaymentType: z.enum(["prepaid", "postpaid"]),
    creditMin: z.string().regex(/^[0-9]+$/, "Credit limit should be a number greater than 0"),
    dueInDays: z.string().regex(/^[0-9]+$/, "Due days should be a number greater than 0"),
  });

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

  const { data: billingDetailsData, isLoading } = useQuery(
    AdminServiceQueries.getBillingAccountDetails,
    create(GetBillingAccountDetailsRequestSchema, {
      orgId: organizationId,
      id: billingId,
    }),
    {
      enabled: !!organizationId && !!billingId,
    },
  );

  useEffect(() => {
    if (billingDetailsData) {
      const data = billingDetailsData;
      const isPostpaid = data?.creditMin < 0n;
      const creditMin = isPostpaid ? data.creditMin * BigInt(-1) : data.creditMin;
      reset({
        tokenPaymentType: isPostpaid ? "postpaid" : "prepaid",
        creditMin: creditMin.toString(),
        dueInDays: data?.dueInDays.toString(),
      });
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

  const { mutateAsync: getBillingAccount } = useMutation(
    FrontierServiceQueries.getBillingAccount,
    {
      onError: (error) => {
        console.error("Unable to fetch billing account:", error);
      },
    },
  );

  const onSubmit = async (data: BillingDetailsForm) => {
    // Transform data for API submission
    const creditMinValue = BigInt(data.creditMin);
    const creditMinForApi = data.tokenPaymentType === "postpaid"
      ? -creditMinValue
      : creditMinValue;

    await updateBillingDetails(
      create(UpdateBillingAccountDetailsRequestSchema, {
        orgId: organizationId,
        id: billingId,
        creditMin: creditMinForApi,
        dueInDays: BigInt(data.dueInDays),
      }),
    );

    const getBillingResp = await getBillingAccount(
      create(GetBillingAccountRequestSchema, {
        orgId: organizationId,
        id: billingId,
        withBillingDetails: true,
      }),
    );

    const updatedDetails = getBillingResp?.billingDetails;
    if (updatedDetails && setBillingAccountDetails) {
      setBillingAccountDetails({
        credit_min: updatedDetails.creditMin.toString(),
        due_in_days: updatedDetails.dueInDays.toString(),
      });
    }
    toast.success("Billing details updated");
  };

  const onValueChange = (value: string) => {
    const paymentType = value as BillingDetailsForm["tokenPaymentType"];
    setValue("tokenPaymentType", paymentType);
    if (paymentType === "prepaid") {
      setValue("creditMin", "0");
      setValue("dueInDays", "0");
    } else {
      setValue("creditMin", "0");
      setValue("dueInDays", "30");
    }
  };

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
                  type="string"
                  {...register("creditMin", { pattern: { value: /^\d+$/, message: "Credit limit must be a number" } })}
                  error={errors?.creditMin?.message}
                />
              )}
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <InputField
                  disabled={isPrepaid}
                  label="Billing due date"
                  type="string"
                  suffix="Days"
                  {...register("dueInDays", { pattern: { value: /^\d+$/, message: "due days must be in number" } })}
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
