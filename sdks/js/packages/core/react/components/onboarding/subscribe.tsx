'use client';

import { Flex, InputField, Image, Text, Button } from '@raystack/apsara/v1';
import styles from './onboarding.module.css';
import { useSearch } from '@tanstack/react-router';
import PixxelLogoMonogram from '~/react/assets/logos/pixxel-logo-monogram.svg';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { useForm } from 'react-hook-form';
import { ReactNode } from 'react';

type SubscribeProps = {
  logo?: string | ReactNode;
  preferenceTitle?: string;
  preferenceDescription?: string;
  onSubmit?: (data: FormData) => void;
};

const schema = yup.object({
  name: yup.string().required('Name is required'),
  email: yup.string().email('Invalid email address').required('Email is required'),
  contact: yup.string().optional().matches(/^\+\d+$/, 'Contact must include country code'),
  activity: yup.string().required('Activity is required'),
  utm_medium: yup.string().optional(),
  utm_source: yup.string().optional(),
});

type FormData = yup.InferType<typeof schema>;

type SearchParams = {
  activity?: string;
  utm_medium?: string;
  utm_source?: string;
  title?: string;
  desc?: string;
};

export const Subscribe = ({
  logo = PixxelLogoMonogram as unknown as string,
  preferenceTitle = 'Updates, News & Events',
  preferenceDescription = 'Stay informed on new features, improvements, and key updates',
  onSubmit
}: SubscribeProps) => {
  const search = useSearch({ strict: false }) as SearchParams;
  const activity = search.activity || '';
  const utm_medium = search.utm_medium;
  const utm_source = search.utm_source;
  const title = search.title || preferenceTitle;
  const desc = search.desc || preferenceDescription;

  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: yupResolver(schema),
    defaultValues: {
      activity,
    }
  });

  const onFormSubmit = async (data: FormData) => {
    console.log('Form data submitted:', data);

    const requestData: any = {
      name: data.name,
      email: data.email,
      contact: data.contact,
      activity: data.activity,
    };

    if (utm_medium) {
      requestData.metadata = { medium: utm_medium };
    }

    if (utm_source) {
      requestData.source = utm_source;
    }

    // Call the API with requestData
    // Example: await apiCall(requestData);

    onSubmit?.(data);
  };

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
          <Flex direction="column" gap="small">
            <Text size={6} className={styles.subscribeTitle}>{title}</Text>
            <Text size={4} className={styles.subscribeDesc}>{desc}</Text>
          </Flex>

          <InputField 
            label='Name' 
            {...register('name')} 
            error={errors.name?.message}
            ref={undefined}
          />
          <InputField 
            label='Email' 
            {...register('email')} 
            error={errors.email?.message}
            ref={undefined}
          />
          <InputField 
            label='Contact Number' 
            helperText='Add country code at the start' 
            {...register('contact')} 
            error={errors.contact?.message}
            ref={undefined}
          />
          <Button type="submit" width="100%" data-test-id="sdk-demo-subscribe-button">Subscribe</Button>
        </Flex>
      </form>
    </Flex>
  );
};
