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
    .matches(/^\d+$/, 'Must be only digits')
    .min(10, 'Contact number must be at least 10 digits')
    .required('Contact number is required')
});

type FormData = yup.InferType<typeof schema>;

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

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const titleFromQuery = searchParams.get('title');
    const descriptionFromQuery = searchParams.get('description');

    if (titleFromQuery) setTitle(decodeURIComponent(titleFromQuery));
    if (descriptionFromQuery) setDescription(decodeURIComponent(descriptionFromQuery));
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
                label="Name"
                {...register('name')}
                placeholder="Enter your name"
                error={errors.name?.message}
                ref={undefined}
            />

            
            <InputField
                label="Email"
                type="email"
                placeholder="Enter your email"
                error={errors.email?.message}
                {...register('email')}
                ref={undefined}
            />
            
            <InputField
                label="Contact number"
                placeholder="Enter your contact with country code"
                error={errors.contactNumber?.message}
                helperText='Add country code at the start'
                {...register('contactNumber')}
                ref={undefined}
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
