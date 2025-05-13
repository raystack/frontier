import { yupResolver } from '@hookform/resolvers/yup';
import { Button } from '@raystack/apsara/v1';
import { Flex, Separator, Text, TextField } from '@raystack/apsara';
import React, { useCallback, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import isEmail from 'validator/lib/isEmail';
import { HttpErrorResponse } from '~/react/utils';

const styles = {
  container: {
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    gap: 'var(--pd-16)'
  },

  button: {
    width: '100%'
  }
};

type MagicLinkProps = {
  open?: boolean;
  children?: React.ReactNode;
};

const emailSchema = yup.object({
  email: yup
    .string()
    .trim()
    .email()
    .required()
    .test(
      'is-valid',
      message => `${message.path} is invalid`,
      value =>
        value ? isEmail(value) : new yup.ValidationError('Invalid value')
    )
});

type FormData = yup.InferType<typeof emailSchema>;

export const MagicLink = ({ open = false, ...props }: MagicLinkProps) => {
  const { client, config } = useFrontier();
  const [visiable, setVisiable] = useState<boolean>(open);
  const [loading, setLoading] = useState<boolean>(false);

  const {
    watch,
    control,
    handleSubmit,
    setError,
    formState: { errors }
  } = useForm({
    resolver: yupResolver(emailSchema)
  });

  const magicLinkHandler = useCallback(
    async (data: FormData) => {
      setLoading(true);
      try {
        if (!client) return;

        const {
          data: { state = '' }
        } = await client.frontierServiceAuthenticate('mailotp', {
          email: data.email,
          callback_url: config.callbackUrl
        });

        const searchParams = new URLSearchParams({ state, email: data.email });

        // @ts-ignore
        window.location = `${
          config.redirectMagicLinkVerify
        }?${searchParams.toString()}`;
      } catch (err: unknown) {
        if (err instanceof Response && err?.status === 400) {
          const message =
            (err as HttpErrorResponse)?.error?.message || 'Bad Request';
          setError('email', { message });
        } else {
          setError('email', { message: 'An unexpected error occurred' });
        }
      } finally {
        setLoading(false);
      }
    },
    [client, config.callbackUrl, config.redirectMagicLinkVerify, setError]
  );

  const email = watch('email', '');

  if (!visiable)
    return (
      <Button
        variant="solid"
        color="neutral"
        style={styles.button}
        onClick={() => setVisiable(true)}
        data-test-id="frontier-sdk-mail-otp-login-btn"
      >
        Continue with Email
      </Button>
    );

  return (
    <form
      style={{ ...styles.container, flexDirection: 'column' }}
      onSubmit={handleSubmit(magicLinkHandler)}
    >
      {!open && <Separator />}
      <Flex
        direction={'column'}
        align={'start'}
        style={{
          width: '100%',
          position: 'relative',
          marginBottom: 'var(--pd-16)'
        }}
      >
        <Controller
          render={({ field }) => (
            <TextField
              {...field}
              // @ts-ignore
              size="medium"
              placeholder="name@example.com"
            />
          )}
          control={control}
          name="email"
        />

        <Text
          size={1}
          style={{
            color: 'var(--foreground-danger)',
            position: 'absolute',
            top: 'calc(100% + 4px)'
          }}
        >
          {errors.email && String(errors.email?.message)}
        </Text>
      </Flex>
      <Button
        {...props}
        style={{ ...styles.button }}
        disabled={!email}
        type="submit"
        loading={loading}
        loaderText="Loading..."
        data-test-id="frontier-sdk-mail-otp-login-submit-btn"
      >
        Continue with Email
      </Button>
    </form>
  );
};
