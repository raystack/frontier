'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { ReactNode } from '@tanstack/react-router';
import { Button, Flex, Text, Switch, Skeleton } from '@raystack/apsara/v1';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { PREFERENCE_OPTIONS } from '~/react/utils/constants';
import { usePreferences } from '~/react/hooks/usePreferences';
import { Container } from '../Container';
import { Header } from '../Header';
import styles from './onboarding.module.css';

const schema = yup.object({
  [PREFERENCE_OPTIONS.NEWSLETTER]: yup.boolean().optional()
});

type FormData = yup.InferType<typeof schema>;

type UpdatesProps = {
  logo?: ReactNode;
  title?: string;
  preferenceTitle?: string;
  preferenceDescription?: string;
  onSubmit?: (data: FormData) => void;
};

export const Updates = ({
  logo,
  title = 'Subscribe for updates',
  preferenceTitle = 'Updates, News & Events',
  preferenceDescription = 'Stay informed on new features, improvements, and key updates',
  onSubmit
}: UpdatesProps) => {
  const { preferences, isFetching, updatePreferences } = usePreferences();

  const newsletterValue =
    preferences?.[PREFERENCE_OPTIONS.NEWSLETTER]?.value === 'true';

  const {
    control,
    handleSubmit,
    formState: { isSubmitting }
  } = useForm<FormData>({
    values: {
      [PREFERENCE_OPTIONS.NEWSLETTER]: newsletterValue
    },
    resolver: yupResolver(schema)
  });

  async function onFormSubmit(data: FormData) {
    return updatePreferences([
      {
        name: PREFERENCE_OPTIONS.NEWSLETTER,
        value: String(data[PREFERENCE_OPTIONS.NEWSLETTER])
      }
    ])
      .then(() => onSubmit?.(data))
      .catch(err => {
        console.error('frontier:sdk:: error during submit', err);
      });
  }
  return (
    <Flex direction="column" gap="large">
      <Header logo={logo} title={title} />
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <Container
          className={styles.updatesContainer}
          shadow="sm"
          radius="xs"
        >
          <Flex direction="column" gap="medium">
            <Flex justify="between">
              <Text size={6} weight={500}>
                {preferenceTitle}
              </Text>
              {isFetching ? (
                <Skeleton width={34} height={20} />
              ) : (
                <Controller
                  render={({ field: { value, onChange, ...field } }) => (
                    <Switch
                      checked={value}
                      onCheckedChange={onChange}
                      {...field}
                    />
                  )}
                  control={control}
                  name={PREFERENCE_OPTIONS.NEWSLETTER}
                />
              )}
            </Flex>
            <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
              {preferenceDescription}
            </Text>
          </Flex>
          <Button
            style={{ width: '100%' }}
            type="submit"
            data-test-id="frontier-sdk-updates-btn"
            disabled={isFetching || isSubmitting}
            loading={isSubmitting}
            loaderText="Loading..."
          >
            Continue
          </Button>
        </Container>
      </form>
    </Flex>
  );
};
