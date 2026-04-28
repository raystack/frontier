import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  Text,
  Separator,
  Flex,
  InputField
} from '@raystack/apsara-v1';
import { ReactNode, useCallback, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import isEmail from 'validator/lib/isEmail';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { HttpErrorResponse } from '~/react/utils';
import styles from './magic-link-form.module.css';

type MagicLinkFormProps = {
  open?: boolean;
  children?: ReactNode;
};

const emailSchema = yup.object({
  email: yup
    .string()
    .trim()
    .required()
    .test(
      'is-valid',
      () => 'Please enter a valid email address.',
      value =>
        value ? isEmail(value) : new yup.ValidationError('Invalid value')
    )
});

type FormData = yup.InferType<typeof emailSchema>;

export const MagicLinkForm = ({
  open = false,
  ...props
}: MagicLinkFormProps) => {
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
        className={styles.button}
        onClick={() => setVisible(true)}
        data-test-id="frontier-sdk-mail-otp-login-btn"
      >
        Continue with Email
      </Button>
    );

  return (
    <form className={styles.form} onSubmit={handleSubmit(magicLinkHandler)}>
      {!open && <Separator />}
      <Flex direction="column" align="start" className={styles.field}>
        <InputField
          {...register('email')}
          size="large"
          placeholder="name@example.com"
        />
        <Text size="mini" variant="danger" className={styles.error}>
          {errors.email && String(errors.email?.message)}
        </Text>
      </Flex>
      <Button
        {...props}
        className={styles.button}
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
