import { useMemo } from 'react';
import { Button, Dialog, Flex, InputField, Skeleton } from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { yupResolver } from '@hookform/resolvers/yup';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_TOKEN_PRODUCT_NAME } from '~/react/utils/constants';
import { CreateCheckoutRequestSchema, FrontierServiceQueries } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import qs from 'query-string';

type DialogHandle = ReturnType<typeof Dialog.createHandle>;

export interface AddTokensDialogProps {
  handle: DialogHandle;
}

export function AddTokensDialog({ handle }: AddTokensDialogProps) {
  const { config, activeOrganization } = useFrontier();

  const tokenProductId =
    config?.billing?.tokenProductId || DEFAULT_TOKEN_PRODUCT_NAME;

  const { data: product, isLoading } = useQuery(
    FrontierServiceQueries.getProduct,
    { id: tokenProductId },
    {
      enabled: !!tokenProductId,
      select: data => data?.product
    }
  );

  const { productDescription, minQuantity, maxQuantity } = useMemo(() => {
    let productDescription = '';
    let minQuantity = 1;
    let maxQuantity = 1000000;
    if (product) {
      const productPrice = product?.prices?.[0];
      const price = Number(productPrice?.amount || '100') / 100;
      const currency = productPrice?.currency || 'USD';
      productDescription = `1 token = ${currency} ${price}`;
    }
    const behaviorConfig = product?.behaviorConfig;
    if (behaviorConfig?.minQuantity) {
      minQuantity = Number(behaviorConfig?.minQuantity);
    }
    if (behaviorConfig?.maxQuantity) {
      maxQuantity = Number(behaviorConfig?.maxQuantity);
    }
    return { productDescription, minQuantity, maxQuantity };
  }, [product]);

  const tokensSchema = yup
    .object({
      tokens: yup
        .number()
        .required('Please enter valid number')
        .min(minQuantity, `Minimum ${minQuantity} token is required`)
        .max(maxQuantity, `Maximum ${maxQuantity} tokens are allowed`)
        .typeError('Please enter valid number of tokens')
    })
    .required();

  type FormData = yup.InferType<typeof tokensSchema>;

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setValue
  } = useForm({
    resolver: yupResolver(tokensSchema)
  });

  const { mutateAsync: createCheckout, isPending: isCreatingCheckout } =
    useMutation(FrontierServiceQueries.createCheckout, {
      onSuccess: data => {
        const checkoutUrl = data?.checkoutSession?.checkoutUrl;
        if (checkoutUrl) window.location.href = checkoutUrl;
      },
      onError: (error: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: error?.message,
          type: 'error'
        });
      }
    });

  const onSubmit = async (data: FormData) => {
    try {
      if (!activeOrganization?.id) return;
      const query = qs.stringify(
        {
          details: btoa(
            qs.stringify({
              organization_id: activeOrganization?.id,
              type: 'tokens'
            })
          ),
          checkout_id: '{{.CheckoutID}}'
        },
        { encode: false }
      );
      const cancelUrl = `${config?.billing?.cancelUrl}?${query}`;
      const successUrl = `${config?.billing?.successUrl}?${query}`;

      await createCheckout(
        create(CreateCheckoutRequestSchema, {
          orgId: activeOrganization?.id || '',
          cancelUrl: cancelUrl,
          successUrl: successUrl,
          productBody: {
            product: tokenProductId,
            quantity: BigInt(data.tokens)
          }
        })
      );
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error';
      toastManager.add({
        title: 'Something went wrong',
        description: message,
        type: 'error'
      });
    }
  };

  const isFormSubmitting = isSubmitting || isCreatingCheckout;

  return (
    <Dialog handle={handle}>
      <Dialog.Content width={400} showCloseButton={false}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Dialog.Title>Add tokens</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              {isLoading ? (
                <Skeleton height="60px" width="100%" />
              ) : (
                <InputField
                  label="Add tokens"
                  size="large"
                  type="number"
                  error={errors.tokens && String(errors.tokens.message)}
                  {...register('tokens', { valueAsNumber: true })}
                  placeholder="Enter no. of tokens"
                  helperText={productDescription}
                  onKeyDown={(e: React.KeyboardEvent) =>
                    ['e', 'E', '+', '-', '.'].includes(e.key) &&
                    e.preventDefault()
                  }
                  onPaste={(e: React.ClipboardEvent) => {
                    const pastedText = e.clipboardData.getData('text/plain');
                    const parsedValue = parseInt(pastedText);
                    e.preventDefault();
                    if (
                      !isNaN(parsedValue) &&
                      parsedValue >= minQuantity &&
                      parsedValue <= maxQuantity
                    ) {
                      setValue('tokens', parsedValue, { shouldDirty: true });
                    }
                  }}
                  data-test-id="frontier-sdk-add-tokens-input"
                />
              )}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Flex gap={4} justify="end">
              <Button
                variant="outline"
                color="neutral"
                onClick={() => handle.close()}
                data-test-id="frontier-sdk-add-tokens-cancel-btn"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="solid"
                color="accent"
                loading={isFormSubmitting}
                disabled={!!errors.tokens || isFormSubmitting || isLoading}
                loaderText="Adding..."
                data-test-id="frontier-sdk-add-tokens-btn"
              >
                Continue
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
