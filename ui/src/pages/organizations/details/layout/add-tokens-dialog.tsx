import {
  Button,
  Dialog,
  Flex,
  InputField,
  Label,
  Text,
  toast,
} from "@raystack/apsara";
import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { Controller, FormProvider, useForm } from "react-hook-form";
import styles from "./layout.module.css";
import * as z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { AppContext } from "~/contexts/App";
import { defaultConfig } from "~/utils/constants";
import { useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AdminServiceQueries, CheckoutProductBodySchema, DelegatedCheckoutRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

interface InviteUsersDialogProps {
  onOpenChange: (open: boolean) => void;
}

const addTokensSchema = z.object({
  product: z.string(),
  quantity: z.coerce
    .number({
      required_error: "Quantity is required",
      invalid_type_error: "Please enter a valid number",
    })
    .min(1)
    .transform((num) => num.toString()),
});

type AddTokenRequestType = z.infer<typeof addTokensSchema>;

export const AddTokensDialog = ({ onOpenChange }: InviteUsersDialogProps) => {
  const { config } = useContext(AppContext);
  const { organization, billingAccount, fetchTokenBalance } =
    useContext(OrganizationContext);
  const queryClient = useQueryClient();
  const transport = useTransport();
  const organisationId = organization?.id || "";
  const billingAccountId = billingAccount?.id || "";

  const methods = useForm<AddTokenRequestType>({
    resolver: zodResolver(addTokensSchema),
    defaultValues: {
      quantity: "0",
      product: config?.token_product_id || defaultConfig.token_product_id,
    },
  });

  const { mutateAsync: delegatedCheckout } = useMutation(
    AdminServiceQueries.delegatedCheckout,
    {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: AdminServiceQueries.searchOrganizationTokens,
            transport,
            input: { id: organisationId },
            cardinality: "infinite",
          }),
        });
        fetchTokenBalance();
        toast.success("Tokens added");
        onOpenChange(false);
      },
      onError: (error) => {
        toast.error("Something went wrong", {
          description: error.message,
        });
        console.error("Unable to add tokens:", error);
      },
    },
  );

  const onSubmit = async (product_body: AddTokenRequestType) => {
    if (!organisationId) return;
    await delegatedCheckout(
      create(DelegatedCheckoutRequestSchema, {
        orgId: organisationId,
        billingId: billingAccountId,
        productBody: create(CheckoutProductBodySchema, {
          product: product_body.product,
          quantity: BigInt(product_body.quantity),
        }),
      }),
    );
  };

  const isSubmitting = methods?.formState?.isSubmitting;
  const errors = methods?.formState?.errors;

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <Dialog.Content width={400}>
        <FormProvider {...methods}>
          <form onSubmit={methods.handleSubmit(onSubmit)}>
            <Dialog.Header>
              <Dialog.Title>Add tokens</Dialog.Title>
              <Dialog.CloseButton data-test-id="add-tokens-close-button" />
            </Dialog.Header>
            <Dialog.Body>
              <Flex direction="column" gap={7}>
                <Flex direction="column" gap={2}>
                  <Label>Tokens</Label>
                  <Controller
                    name="quantity"
                    control={methods.control}
                    render={({ field }) => {
                      return (
                        <InputField
                          {...field}
                          type="number"
                          min={0}
                          className={styles["add-token-dialog-tokens-field"]}
                          onKeyDown={(e) =>
                            ["+", "-", ".", "e", "E"].includes(e.key) &&
                            e.preventDefault()
                          }
                          onPaste={(e) => e.preventDefault()}
                        />
                      );
                    }}
                  />

                  {errors?.quantity?.message ? (
                    <Text size={1} className={styles["form-error-message"]}>
                      {errors?.quantity?.message}
                    </Text>
                  ) : null}
                </Flex>
              </Flex>
            </Dialog.Body>
            <Dialog.Footer>
              <Dialog.Close asChild>
                <Button
                  data-test-id="add-tokens-invite-button"
                  type="reset"
                  color="neutral"
                  variant="outline"
                >
                  Cancel
                </Button>
              </Dialog.Close>
              <Button
                data-test-id="add-tokens-invite-button"
                type="submit"
                loaderText="Adding..."
                loading={isSubmitting}
              >
                Add
              </Button>
            </Dialog.Footer>
          </form>
        </FormProvider>
      </Dialog.Content>
    </Dialog>
  );
};
