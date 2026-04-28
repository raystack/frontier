'use client';

import { yupResolver } from '@hookform/resolvers/yup';
import { type ReactNode } from 'react';
import { Button, Flex, Text, Switch, Skeleton } from '@raystack/apsara-v1';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { PREFERENCE_OPTIONS } from '~/react/utils/constants';
import { usePreferences } from '~/react/hooks/usePreferences';
import { AuthContainer } from '../components/auth-container';
import { AuthHeader } from '../components/auth-header';
import styles from './updates-view.module.css';

const schema = yup.object({
  [PREFERENCE_OPTIONS.NEWSLETTER]: yup.boolean().optional()
});

type FormData = yup.InferType<typeof schema>;

export type UpdatesViewProps = {
  logo?: ReactNode;
  title?: string;
  preferenceTitle?: string;
  preferenceDescription?: string;
  // eslint-disable-next-line no-unused-vars
  onSubmit?: (data: FormData) => void;
};

export const UpdatesView = ({
  logo,
  title = 'Subscribe for updates',
  preferenceTitle = 'Updates, News & Events',
  preferenceDescription = 'Stay informed on new features, improvements, and key updates',
  onSubmit
}: UpdatesViewProps) => {
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
    <Flex direction="column" gap={9}>
      <AuthHeader logo={logo} title={title} />
      <form onSubmit={handleSubmit(onFormSubmit)}>
        <AuthContainer className={styles.updatesContainer}>
          <Flex direction="column" gap={5}>
            <Flex justify="between">
              <Text size="large" weight="medium">
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
            <Text size="regular" variant="secondary">
              {preferenceDescription}
            </Text>
          </Flex>
          <Button
            className={styles.button}
            type="submit"
            data-test-id="frontier-sdk-updates-btn"
            disabled={isFetching || isSubmitting}
            loading={isSubmitting}
            loaderText="Loading..."
          >
            Continue
          </Button>
        </AuthContainer>
      </form>
    </Flex>
  );
};
