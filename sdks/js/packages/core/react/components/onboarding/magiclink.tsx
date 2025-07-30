import { yupResolver } from '@hookform/resolvers/yup';
import { Button, Text, Separator, Flex, InputField } from '@raystack/apsara/v1';
import React, { useCallback, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import isEmail from 'validator/lib/isEmail';
import { HttpErrorResponse } from '~/react/utils';
import { useMutation, FrontierServiceQueries } from '~hooks';

const styles = {
  container: {
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    gap: 'var(--rs-space-5)'
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
  const { config } = useFrontier();
  const [visible, setVisible] = useState<boolean>(open);

  const { mutateAsync: authenticate, isPending } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const {
    watch,
    handleSubmit,
    setError,
    register,
    formState: { errors }
  } = useForm({
    resolver: yupResolver(emailSchema)
  });

  const magicLinkHandler = useCallback(
    async (data: FormData) => {
      try {
        const response = await authenticate({
          strategyName: 'mailotp',
          email: data.email,
          callbackUrl: config.callbackUrl
        });

        const searchParams = new URLSearchParams({
          state: response.state || '',
          email: data.email
        });

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
      }
    },
    [authenticate, config.callbackUrl, config.redirectMagicLinkVerify, setError]
  );

  const email = watch('email', '');

  if (!visible)
    return (
      <Button
        variant="outline"
        color="neutral"
        style={styles.button}
        onClick={() => setVisible(true)}
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
        direction="column"
        align="start"
        style={{
          width: '100%',
          position: 'relative',
          marginBottom: 'var(--rs-space-5)'
        }}
      >
        <InputField
          {...register('email')}
          size="large"
          placeholder="name@example.com"
        />
        <Text
          size="mini"
          variant="danger"
          style={{
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
        loading={isPending}
        loaderText="Loading..."
        data-test-id="frontier-sdk-mail-otp-login-submit-btn"
      >
        Continue with Email
      </Button>
    </form>
  );
};
