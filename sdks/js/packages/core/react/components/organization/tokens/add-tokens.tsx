import {
  Button,
  Image,
  Text,
  Flex,
  toast,
  Dialog,
  InputField,
  Skeleton
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';
import tokenStyles from './token.module.css';
import qs from 'query-string';
import { DEFAULT_TOKEN_PRODUCT_NAME } from '~/react/utils/constants';
import { useState, useEffect } from 'react';

export const AddTokens = () => {
  const [productDescription, setProductDescription] = useState<string>('');
  const [minQuantity, setMinQuantity] = useState<number>(1);
  const [maxQuantity, setMaxQuantity] = useState<number>(1000000);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const navigate = useNavigate({ from: '/tokens/modal' });
  const { config, client, activeOrganization, billingAccount } = useFrontier();

  // Create schema dynamically based on product config
  const tokensSchema = yup
    .object({
      tokens: yup
        .number()
        .required()
        .min(minQuantity, `Minimum ${minQuantity} token is required`)
        .max(maxQuantity, `Maximum ${maxQuantity} tokens are allowed`)
    })
    .typeError('Please enter a valid number')
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

  // Fetch product description for helper text
  useEffect(() => {
    const fetchProductDescription = async () => {
      try {
        const tokenProductId = config?.billing?.tokenProductId || DEFAULT_TOKEN_PRODUCT_NAME;
        const response = await client?.frontierServiceGetProduct(tokenProductId);
        const product = response?.data?.product;

        if (product) {
          // Set price description
          const productPrice = product?.prices?.[0];
          const price = parseInt(productPrice?.amount || "100")/100;
          const currency = productPrice?.currency || "USD";
          const description = `1 token = ${currency} ${price}`;
          setProductDescription(description);

          // Set quantity constraints from behavior config
          const behaviorConfig = product?.behavior_config;
          if (behaviorConfig?.min_quantity) {
            setMinQuantity(parseInt(behaviorConfig.min_quantity));
          }
          if (behaviorConfig?.max_quantity) {
            setMaxQuantity(parseInt(behaviorConfig.max_quantity));
          }
        }
        setIsLoading(false);
      } catch (error) {
        console.error('Failed to fetch product description:', error);
      }
    };

    fetchProductDescription();
  }, [client, config?.billing?.tokenProductId]);

  const onSubmit = async (data: FormData) => {
    if (!client) return;
    if (!activeOrganization?.id) return;

    try {
      if (activeOrganization?.id && billingAccount?.id) {
        // Token product id or name can be used here
        const tokenProductId =
          config?.billing?.tokenProductId || DEFAULT_TOKEN_PRODUCT_NAME;
        const query = qs.stringify(
          {
            details: btoa(
              qs.stringify({
                billing_id: billingAccount?.id,
                organization_id: activeOrganization?.id,
                type: 'tokens'
              })
            ),
            checkout_id: '{{.CheckoutID}}'
          },
          { encode: false }
        );
        const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
        const success_url = `${config?.billing?.successUrl}?${query}`;

        const resp = await client?.frontierServiceCreateCheckout(
          activeOrganization?.id,
          billingAccount?.id,
          {
            cancel_url: cancel_url,
            success_url: success_url,
            product_body: {
              product: tokenProductId,
              quantity: data.tokens.toString()
            }
          }
        );
        if (resp?.data?.checkout_session?.checkout_url) {
          window.location.href = resp?.data?.checkout_session.checkout_url;
        }
      }
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  };

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
              {isLoading ? <Skeleton count={3}/> :
              <InputField
                label="Add tokens"
                size="large"
                type="number"
                error={errors.tokens && String(errors.tokens.message)}
                {...register('tokens', {valueAsNumber: true})}
                name="tokens"
                placeholder="Enter no. of tokens"
                helperText={productDescription}
                className={tokenStyles.tokenInputField}
                onKeyDown={(e) => ['e','E','+','-','.'].includes(e.key) && e.preventDefault()}
                onPaste={(e) => {
                  const pastedText = e.clipboardData.getData('text/plain');
                  const parsedValue = parseInt(pastedText);
                  e.preventDefault();
                  if (!isNaN(parsedValue) && parsedValue >= minQuantity && parsedValue <= maxQuantity) {
                    setValue('tokens', parsedValue);
                  }
                }}
              />
              }
            </Flex>
          </Dialog.Body>

          <Dialog.Footer>
            <Button variant="outline" color="neutral" onClick={() => navigate({ to: '/tokens' })}>Cancel</Button>
            <Button
              type="submit"
              loading={isSubmitting}
              disabled={!!errors.tokens || isSubmitting || isLoading}
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
