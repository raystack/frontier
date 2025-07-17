'use client';

import { ReactNode, useState } from 'react';
import * as yup from 'yup';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';

import { Button, Flex, Text, InputField, toast, ToastContainer, Image, EmptyState } from '@raystack/apsara/v1';
import { useFrontier } from '~/react/contexts/FrontierContext';

import checkCircle from '~/react/assets/check-circle.svg';
import styles from './onboarding.module.css';

const schema = yup.object({
  name: yup.string().required('Name is required'),
  email: yup.string().email('Invalid email').required('Email is required'),
  contactNumber: yup
    .string()
    .transform((value) => value.trim() === '' ? null : value)
    .nullable()
    .test('digits-only', 'Must contain only numbers with country code', (value) => {
      if (!value?.trim()) return true;
      return /^[+\d\s\-()]+$/.test(value);
    })
    .optional()
});

type FormData = yup.InferType<typeof schema>;

interface ExtendedFormData extends FormData {
  activity: string;
  source?: string;
  metadata?: {
    medium?: string;
  };
}

type SubscribeProps = {
  title?: string;
  desc?: string;
  activity?: string;
  medium?: string;
  source?: string;
  confirmSection?: ReactNode;
  // eslint-disable-next-line no-unused-vars
  onSubmit?: (data: FormData) => void;
};

const DEFAULT_TITLE = 'Updates, News & Events';
const DEFAULT_DESCRIPTION = 'Stay informed on new features, improvements, and key updates';
const DEFAULT_SUCCESS_TITLE = 'Thank you for subscribing!';
const DEFAULT_SUCCESS_DESCRIPTION = 'You have successfully subscribed to our list. We will let you know about the updates.';

const ConfirmSection = () => {
  return (
    <Flex direction="column" gap={9} align="center" justify="center">
        <EmptyState
          icon={<Image alt="" width={32} height={32} src={checkCircle as unknown as string} />}
          heading={DEFAULT_SUCCESS_TITLE}
          subHeading={DEFAULT_SUCCESS_DESCRIPTION}
        />
        <ToastContainer />
      </Flex>
  );
};

export const Subscribe = ({
  title = DEFAULT_TITLE,
  desc = DEFAULT_DESCRIPTION,
  activity = 'newsletter',
  medium,
  source,
  confirmSection = <ConfirmSection />,
  onSubmit
}: SubscribeProps) => {
  const { client } = useFrontier();
  const [isSuccess, setIsSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<FormData>({
    resolver: yupResolver(schema)
  });

  async function onFormSubmit(data: FormData) {
    try {
      const formData: ExtendedFormData = { ...data, activity };
      if (medium) {
        formData.metadata = { ...formData.metadata, medium };
      }
      if (source) {
        formData.source = source;
      }

      const response = await client?.frontierServiceCreateProspectPublic({
        name: formData.name,
        email: formData.email,
        phone: formData?.contactNumber || undefined,
        activity: formData.activity,
        source: formData.source,
        metadata: formData.metadata
      });

      if (response?.status === 200) {
        setIsSuccess(true);
      }

      await onSubmit?.(data);
    } catch (err) {
      console.error('Frontier SDK: Error while submitting the form', err);
      toast.error('Something went wrong. Please try again.');
      throw err;
    }
  }

  if (isSuccess) {
    return (
      <>
        {confirmSection}
      </>
    );
  }

  return (
    <Flex direction="column" gap={9} align="center" justify="center">
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <Flex
          className={styles.subscribeContainer}
          direction='column'
          justify='start'
          align="start"
          gap={9}
        >
          <Flex direction="column" gap={3} style={{ width: '100%' }}>
            <Text size="large" className={styles.subscribeTitle}>{title}</Text>
            <Text size="regular" className={styles.subscribeDescription}>{desc}</Text>
          </Flex>
          <InputField
            {...register('name')}
            label="Name"
            placeholder="Enter name"
            error={errors.name?.message}
            data-testid="subscribe-name-input"
          />
          <InputField
            {...register('email')}
            label="Email"
            type="email"
            placeholder="Enter email"
            error={errors.email?.message}
            data-testid="subscribe-email-input"
          />
          <InputField
            {...register('contactNumber')}
            optional
            label="Contact number"
            placeholder="Enter contact"
            error={errors.contactNumber?.message}
            helperText='Add country code at the start'
            data-testid="subscribe-contact-input"
          />
          <Button
            style={{ width: '100%' }}
            type="submit"
            data-test-id="frontier-sdk-subscribe-btn"
            disabled={isSubmitting}
            loading={isSubmitting}
            loaderText="Submitting..."
          >
            Subscribe
          </Button>
        </Flex>
        <ToastContainer />
      </form>
    </Flex>
  );
};
