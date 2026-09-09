import { yupResolver } from '@hookform/resolvers/yup';
import { Button, Separator, Field, Input } from '@raystack/apsara';
import { ComponentPropsWithRef, ReactNode, useCallback, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import isEmail from 'validator/lib/isEmail';
import { useMutation } from '@connectrpc/connect-query';
import { FlowIntent, FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/client/contexts/FrontierContext';
import {
  AuthContainer,
  type AuthContainerProps
} from '~/client/components/auth-container';
import { AuthHeader } from '~/client/components/auth-header';
import { describeAuthError } from '~/client/utils/auth-error';
import { MAIL_OTP_STRATEGY } from '~/client/utils/constants';
import styles from './magic-link-view.module.css';

export type MagicLinkViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    open?: boolean;
    inline?: boolean;
    intent?: FlowIntent;
    acceptedDocumentIds?: string[];
    disabled?: boolean;
    onActivate?: () => void;
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

export const MagicLinkView = ({
  logo,
  title = 'Login to Raystack',
  open = false,
  inline = false,
  intent = FlowIntent.UNSPECIFIED,
  acceptedDocumentIds,
  disabled = false,
  onActivate,
  ...props
}: MagicLinkViewProps) => {
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
      onActivate?.();
      try {
        const response = await authenticate({
          strategyName: MAIL_OTP_STRATEGY,
          email: data.email,
          callbackUrl: config.callbackUrl,
          flowIntent: intent,
          acceptedDocumentIds
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
        setError('email', { message: describeAuthError(err).message });
      }
    },
    [authenticate, config, intent, acceptedDocumentIds, setError, onActivate]
  );

  const email = watch('email', '');

  const formContent = !visible ? (
    <Button
      variant="outline"
      color="neutral"
      className={styles.button}
      onClick={() => {
        setVisible(true);
        onActivate?.();
      }}
      disabled={disabled}
      data-test-id="frontier-sdk-mail-otp-login-btn"
    >
      Continue with Email
    </Button>
  ) : (
    <form
      noValidate
      className={styles.form}
      onSubmit={handleSubmit(magicLinkHandler)}
    >
      {!open && <Separator />}
      <Field error={errors.email?.message}>
        <Input
          {...register('email')}
          size="large"
          placeholder="name@example.com"
          disabled={disabled}
        />
      </Field>
      <Button
        className={styles.button}
        disabled={!email || disabled}
        type="submit"
        loading={isPending}
        loaderText="Loading..."
        data-test-id="frontier-sdk-mail-otp-login-submit-btn"
      >
        Continue with Email
      </Button>
    </form>
  );

  if (inline) return formContent;

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      {formContent}
    </AuthContainer>
  );
};
