'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { Button, Flex, Text, InputField } from '@raystack/apsara/v1';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import styles from './onboarding.module.css';
import PixxelLogoMonogram from '~/react/assets/logos/pixxel-logo-monogram.svg';
import { Image } from '@raystack/apsara/v1';
import { ReactNode, useEffect, useState } from 'react';

const schema = yup.object({
  name: yup.string().required('Name is required'),
  email: yup.string().email('Invalid email').required('Email is required'),
  contactNumber: yup
    .string()
    .transform((value) => value.trim() === '' ? null : value)
    .nullable()
    .test('digits-only', 'Must be only digits', (value) => {
      if (!value?.trim()) return true;
      return /^\d+$/.test(value);
    })
    .optional()
});

type FormData = yup.InferType<typeof schema>;

interface ExtendedFormData extends FormData {
  activity: string;
  status: string;
  source?: string;
  metadata?: {
    medium?: string;
  };
}

type SubscribeProps = {
  logo?: ReactNode;
  title?: string;
  description?: string;
  onSubmit?: (data: FormData) => void;
};

const DEFAULT_TITLE = 'Updates, News & Events';
const DEFAULT_DESCRIPTION = 'Stay informed on new features, improvements, and key updates';

export const Subscribe = ({
  logo = PixxelLogoMonogram as unknown as string,
  title: defaultTitle = DEFAULT_TITLE,
  description: defaultDescription = DEFAULT_DESCRIPTION,
  onSubmit
}: SubscribeProps) => {
  const [title, setTitle] = useState(defaultTitle);
  const [description, setDescription] = useState(defaultDescription);
  const [activity, setActivity] = useState('');
  const [status, setStatus] = useState('');
  const [medium, setMedium] = useState<string | null>(null);
  const [source, setSource] = useState<string | null>(null);

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const titleFromQuery = searchParams.get('title');
    const descriptionFromQuery = searchParams.get('description');
    const activityFromQuery = searchParams.get('activity') || '';
    const statusFromQuery = searchParams.get('status') || 'subscribed';
    const utmMedium = searchParams.get('utm_medium');
    const utmSource = searchParams.get('utm_source');

    if (titleFromQuery) setTitle(decodeURIComponent(titleFromQuery));
    if (descriptionFromQuery) setDescription(decodeURIComponent(descriptionFromQuery));
    if (activityFromQuery) setActivity(decodeURIComponent(activityFromQuery));
    if (statusFromQuery) setStatus(decodeURIComponent(statusFromQuery));
    if (utmMedium) setMedium(decodeURIComponent(utmMedium));
    if (utmSource) setSource(decodeURIComponent(utmSource));
  }, []);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm<FormData>({
    resolver: yupResolver(schema)
  });

  async function onFormSubmit(data: FormData) {
    try {
      const formData: ExtendedFormData = { ...data, activity, status };
      if (medium) {
        formData.metadata = { ...formData.metadata, medium };
      }
      if (source) {
        formData.source = source;
      }
      console.log('data', formData);
      await onSubmit?.(data);
    } catch (err) {
      console.error('frontier:sdk:: error during submit', err);
    }
  }

  return (
    <Flex direction="column" gap="large" align="center" justify="center">
    {typeof logo === 'string' ? (
      <Image alt="" width={88} height={88} src={logo} />
    ) : (
      logo
    )}
    <form onSubmit={handleSubmit(onFormSubmit)}>
      <Flex
        className={styles.subscribeContainer}
        direction='column'
        justify='start'
        align="start"
        gap="large"
      >
        <Flex direction="column" gap="small" style={{ width: '100%' }}>
          <Text size={6} className={styles.subscribeTitle}>{title}</Text>
          <Text size={4} className={styles.subscribeDescription}>{description}</Text>
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
                loaderText="Loading..."
            >
                Subscribe
            </Button>
        </Flex>
      </form>
    </Flex>
  );
};
