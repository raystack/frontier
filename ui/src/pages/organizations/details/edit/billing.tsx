import { Cross1Icon } from "@radix-ui/react-icons";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Sheet,
  SidePanel,
  toast,
} from "@raystack/apsara/v1";
import styles from "./edit.module.css";
import { useCallback, useContext, useEffect, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { api } from "~/api";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Skeleton from "react-loading-skeleton";

interface EditBillingPanelProps {
  onClose: () => void;
}

const billingDetailsUpdateSchema = z.object({
  credit_min: z.number(),
  due_in_days: z.number().min(0),
});

type BillingDetailsForm = z.infer<typeof billingDetailsUpdateSchema>;

export function EditBillingPanel({ onClose }: EditBillingPanelProps) {
  const { billingAccount } = useContext(OrganizationContext);
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const {
    reset,
    register,
    handleSubmit,
    formState: { isSubmitting, errors },
  } = useForm<BillingDetailsForm>({
    resolver: zodResolver(billingDetailsUpdateSchema),
  });

  const getBillingDetails = useCallback(
    async (orgId: string, billingId: string) => {
      setIsLoading(true);
      try {
        const resp = await api?.adminServiceGetBillingAccountDetails(
          orgId,
          billingId,
        );
        const data = resp?.data;
        reset({
          credit_min: Number(data?.credit_min),
          due_in_days: Number(data?.due_in_days),
        });
      } catch (error) {
        console.error("Failed to fetch billing details:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [reset],
  );

  useEffect(() => {
    if (billingAccount?.org_id && billingAccount?.id) {
      getBillingDetails(billingAccount?.org_id, billingAccount?.id);
    }
  }, [billingAccount?.org_id, billingAccount?.id, getBillingDetails]);

  const onSubmit = async (data: BillingDetailsForm) => {
    await api?.adminServiceUpdateBillingAccountDetails(
      billingAccount?.org_id || "",
      billingAccount?.id || "",
      {
        credit_min: data.credit_min.toString(),
        due_in_days: data.due_in_days.toString(),
      },
    );
    toast.success("Billing details updated");
  };

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
                <InputField
                  label="Credit limit"
                  type="number"
                  {...register("credit_min", { valueAsNumber: true })}
                  error={errors?.credit_min?.message}
                />
              )}
              {isLoading ? (
                <Skeleton height={"32px"} />
              ) : (
                <InputField
                  label="Billing due date"
                  type="number"
                  suffix="Days"
                  {...register("due_in_days", { valueAsNumber: true })}
                  error={errors?.due_in_days?.message}
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
