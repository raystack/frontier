import {
  Button,
  Image,
  Text,
  Flex,
  toast,
  Dialog,
  InputField,
  Skeleton
} from '@raystack/apsara';
import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useMutation } from '@connectrpc/connect-query';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';
import tokenStyles from './token.module.css';
import qs from 'query-string';
import { DEFAULT_TOKEN_PRODUCT_NAME } from '~/react/utils/constants';
import { useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { CreateCheckoutRequestSchema, FrontierServiceQueries } from '~/src';
import { create } from '@bufbuild/protobuf';

export const AddTokens = () => {
  const navigate = useNavigate({ from: '/tokens/modal' });
  const { config, activeOrganization, billingAccount } = useFrontier();

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
    // Set quantity constraints from behavior config
    const behaviorConfig = product?.behaviorConfig;
    if (behaviorConfig?.minQuantity) {
      minQuantity = Number(behaviorConfig?.minQuantity);
    }
    if (behaviorConfig?.maxQuantity) {
      maxQuantity = Number(behaviorConfig?.maxQuantity);
    }
    return { productDescription, minQuantity, maxQuantity };
  }, [product]);

  // Create schema dynamically based on product config
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
        toast.error('Something went wrong', {
          description: error?.message
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
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    }
  };

  const isFormSubmitting = isSubmitting || isCreatingCheckout;

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Flex justify="between" align="center" style={{ width: '100%' }}>
              <Text size="large" weight="medium">
                Add tokens
              </Text>

              <Image
                alt="cross"
                style={{ cursor: 'pointer' }}
                src={cross as unknown as string}
                onClick={() => navigate({ to: '/tokens' })}
                data-test-id="frontier-sdk-add-tokens-btn"
              />
            </Flex>
          </Dialog.Header>

          <Dialog.Body>
            <Flex direction="column" gap={5}>
              {isLoading ? (
                <Skeleton count={3} />
              ) : (
                <InputField
                  label="Add tokens"
                  size="large"
                  type="number"
                  error={errors.tokens && String(errors.tokens.message)}
                  {...register('tokens', { valueAsNumber: true })}
                  placeholder="Enter no. of tokens"
                  helperText={productDescription}
                  className={tokenStyles.tokenInputField}
                  onKeyDown={e =>
                    ['e', 'E', '+', '-', '.'].includes(e.key) &&
                    e.preventDefault()
                  }
                  onPaste={e => {
                    const pastedText = e.clipboardData.getData('text/plain');
                    const parsedValue = parseInt(pastedText);
                    e.preventDefault();
                    if (
                      !isNaN(parsedValue) &&
                      parsedValue >= minQuantity &&
                      parsedValue <= maxQuantity
                    ) {
                      setValue('tokens', parsedValue);
                    }
                  }}
                  data-test-id="frontier-sdk-add-tokens-input"
                />
              )}
            </Flex>
          </Dialog.Body>

          <Dialog.Footer>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => navigate({ to: '/tokens' })}
              data-test-id="frontier-sdk-add-tokens-cancel-btn"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={isFormSubmitting}
              disabled={!!errors.tokens || isFormSubmitting || isLoading}
              loaderText="Adding..."
              data-test-id="frontier-sdk-add-tokens-btn"
            >
              Continue
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
