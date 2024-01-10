import {
  Dialog,
  Flex,
  Image,
  Text,
  Separator,
  Button,
  InputField,
  TextField
} from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import styles from '../../organization.module.css';
import cross from '~/react/assets/cross.svg';
import * as yup from 'yup';
import { Controller, FieldPath, useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { toast } from 'sonner';
import { useEffect } from 'react';
import * as _ from 'lodash';

const billingDetailsSchema = yup
  .object({
    name: yup.string().required(),
    email: yup.string().trim().email().required(),
    address: yup.object({
      line1: yup.string(),
      city: yup.string(),
      state: yup.string(),
      postal_code: yup.string(),
      country: yup.string().required('Country is required')
    })
  })
  .required();

type FormData = yup.InferType<typeof billingDetailsSchema>;

interface formField {
  key: Exclude<FieldPath<FormData>, 'address'>;
  label: string;
  placeholder: string;
}

const formFields: formField[] = [
  {
    key: 'name',
    label: 'Name',
    placeholder: 'Please provide name'
  },
  {
    key: 'email',
    label: 'Email',
    placeholder: 'Please provide email'
  },
  {
    key: 'address.country',
    label: 'Country',
    placeholder: 'Please provide country'
  },
  {
    key: 'address.line1',
    label: 'Address',
    placeholder: 'address'
  },
  {
    key: 'address.city',
    label: 'City',
    placeholder: 'city'
  },
  {
    key: 'address.state',
    label: 'State',
    placeholder: 'state'
  },
  {
    key: 'address.postal_code',
    label: 'Pincode',
    placeholder: 'Pincode'
  }
];

export function EditBillingAddress() {
  const navigate = useNavigate({ from: '/billing/$billingId/edit-address' });
  const { billingId } = useParams({ from: '/billing/$billingId/edit-address' });
  const {
    client,
    activeOrganization: organization,
    setBillingAccount,
    billingAccount
  } = useFrontier();

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset
  } = useForm({
    resolver: yupResolver(billingDetailsSchema)
  });

  useEffect(() => {
    if (billingAccount) {
      reset({
        name: billingAccount.name,
        address: billingAccount.address,
        email: billingAccount.email
      });
    }
  }, [billingAccount, reset]);

  async function onSubmit(data: FormData) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      const resp = await client.frontierServiceUpdateBillingAccount(
        organization?.id,
        billingId,
        { body: data }
      );

      if (resp?.data?.billing_account) {
        toast.success('Address updated');
        setBillingAccount(resp?.data?.billing_account);
        navigate({ to: '/billing' });
      }
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const onCancel = () => {
    navigate({ to: '/billing' });
  };

  const isSubmitDisabled = isSubmitting;

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size={6} style={{ fontWeight: '500' }}>
              Billing details
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
              onClick={() => navigate({ to: '/billing' })}
            />
          </Flex>
          <Separator />
          <Flex direction={'column'} gap="medium" style={{ padding: '24px' }}>
            {formFields.map(formField => {
              return (
                <InputField label={formField.label} key={formField.key}>
                  <Controller
                    render={({ field }) => (
                      <TextField
                        {...field}
                        // @ts-ignore
                        size="medium"
                        placeholder={formField.placeholder}
                      />
                    )}
                    control={control}
                    name={formField.key}
                  />

                  <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                    {_.has(errors, formField.key) &&
                      _.chain(errors).get(formField.key).get('message').value()}
                  </Text>
                </InputField>
              );
            })}
          </Flex>
          <Separator />
          <Flex
            justify="end"
            style={{ padding: 'var(--pd-16)' }}
            gap={'medium'}
          >
            <Button variant="secondary" size="medium" onClick={onCancel}>
              Cancel
            </Button>
            <Button
              variant="primary"
              size="medium"
              type="submit"
              disabled={isSubmitDisabled}
            >
              {isSubmitting ? 'Updating...' : 'Update'}
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
