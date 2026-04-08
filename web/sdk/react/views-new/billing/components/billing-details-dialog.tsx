'use client';

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import {
  createConnectQueryKey,
  useMutation
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  Button,
  Dialog,
  Flex,
  InputField,
  Skeleton,
  toastManager
} from '@raystack/apsara-v1';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  UpdateBillingAccountRequestSchema,
  BillingAccountRequestBodySchema,
  BillingAccount_AddressSchema,
  BillingAccount_TaxSchema,
  type BillingAccount
} from '@raystack/proton/frontier';
import { useFrontier } from '../../../contexts/FrontierContext';
import styles from '../billing-view.module.css';

type DialogHandle = ReturnType<typeof Dialog.createHandle>;

const billingDetailsSchema = yup.object({
  name: yup.string().required('Name is required'),
  email: yup.string().email('Invalid email').required('Email is required'),
  address: yup.string().optional(),
  postalCode: yup.string().optional(),
  city: yup.string().optional(),
  state: yup.string().optional(),
  country: yup.string().optional(),
  taxId: yup.string().optional()
});

type BillingDetailsFormData = yup.InferType<typeof billingDetailsSchema>;

export interface BillingDetailsDialogProps {
  handle: DialogHandle;
}

export function BillingDetailsDialog({ handle }: BillingDetailsDialogProps) {
  const { billingAccount, isBillingAccountLoading } = useFrontier();

  return (
    <Dialog handle={handle}>
      {() => (
        <BillingDetailsContent
          handle={handle}
          billingAccount={billingAccount}
          isLoading={isBillingAccountLoading}
        />
      )}
    </Dialog>
  );
}

interface BillingDetailsContentProps {
  handle: DialogHandle;
  billingAccount?: BillingAccount;
  isLoading: boolean;
}

function BillingDetailsContent({
  handle,
  billingAccount,
  isLoading
}: BillingDetailsContentProps) {
  const queryClient = useQueryClient();

  const { mutateAsync: updateBillingAccount, isPending } = useMutation(
    FrontierServiceQueries.updateBillingAccount,
    {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.getBillingAccount,
            cardinality: 'finite'
          })
        });
      }
    }
  );

  const {
    reset,
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<BillingDetailsFormData>({
    resolver: yupResolver(billingDetailsSchema)
  });

  useEffect(() => {
    if (billingAccount) {
      reset({
        name: billingAccount.name || '',
        email: billingAccount.email || '',
        address: billingAccount.address?.line1 || '',
        postalCode: billingAccount.address?.postalCode || '',
        city: billingAccount.address?.city || '',
        state: billingAccount.address?.state || '',
        country: billingAccount.address?.country || '',
        taxId: billingAccount.taxData?.[0]?.id || ''
      });
    }
  }, [billingAccount, reset]);

  async function onSubmit(data: BillingDetailsFormData) {
    if (!billingAccount?.id) return;
    try {
      const taxData = data.taxId
        ? [
            create(BillingAccount_TaxSchema, {
              id: data.taxId,
              type: billingAccount.taxData?.[0]?.type || ''
            })
          ]
        : [];

      await updateBillingAccount(
        create(UpdateBillingAccountRequestSchema, {
          id: billingAccount.id,
          body: create(BillingAccountRequestBodySchema, {
            name: data.name,
            email: data.email,
            address: create(BillingAccount_AddressSchema, {
              line1: data.address || '',
              line2: billingAccount.address?.line2 || '',
              city: data.city || '',
              state: data.state || '',
              postalCode: data.postalCode || '',
              country: data.country || ''
            }),
            taxData
          })
        })
      );

      handle.close();
      toastManager.add({
        title: 'Billing details updated',
        type: 'success'
      });
    } catch (err: unknown) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          err instanceof Error ? err.message : 'Failed to update billing details',
        type: 'error'
      });
    }
  }

  const isBusy = isSubmitting || isPending;

  return (
    <Dialog.Content width={480}>
      <Dialog.Header>
        <Dialog.Title>Billing details</Dialog.Title>
      </Dialog.Header>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Dialog.Body>
          <Flex direction="column" gap={7} className={styles.dialogFormBody}>
            {isLoading ? (
              <>
                <Skeleton height="58px" />
                <Skeleton height="58px" />
                <Skeleton height="58px" />
              </>
            ) : (
              <>
                <InputField
                  label="Name"
                  size="large"
                  {...register('name')}
                  error={errors.name?.message}
                />
                <InputField
                  label="Email"
                  size="large"
                  type="email"
                  {...register('email')}
                  error={errors.email?.message}
                />
                <InputField
                  label="Address"
                  size="large"
                  {...register('address')}
                  error={errors.address?.message}
                />
                <InputField
                  label="Pincode"
                  size="large"
                  {...register('postalCode')}
                  error={errors.postalCode?.message}
                />
                <InputField
                  label="City"
                  size="large"
                  {...register('city')}
                  error={errors.city?.message}
                />
                <InputField
                  label="State"
                  size="large"
                  {...register('state')}
                  error={errors.state?.message}
                />
                <InputField
                  label="Country"
                  size="large"
                  {...register('country')}
                  error={errors.country?.message}
                />
                <InputField
                  label="Tax ID"
                  size="large"
                  optional
                  {...register('taxId')}
                  error={errors.taxId?.message}
                />
              </>
            )}
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              type="button"
              onClick={() => handle.close()}
              data-test-id="frontier-sdk-billing-details-cancel-button"
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="accent"
              type="submit"
              disabled={isLoading || isBusy}
              loading={isBusy}
              loaderText="Updating..."
              data-test-id="frontier-sdk-billing-details-update-button"
            >
              Update
            </Button>
          </Flex>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  );
}
