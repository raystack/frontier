'use client';

import { Button, Text, Link, Flex, InputField } from '@raystack/apsara-v1';
import {
  ChangeEvent,
  ComponentPropsWithRef,
  FormEvent,
  ReactNode,
  useCallback,
  useEffect,
  useRef,
  useState
} from 'react';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  AuthContainer,
  type AuthContainerProps
} from '../components/auth-container';
import { AuthHeader } from '../components/auth-header';
import styles from './magic-link-verify-view.module.css';

export type MagicLinkVerifyViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    redirectURL?: string;
  };

export const MagicLinkVerifyView = ({
  logo,
  title = 'Check your email',
  redirectURL,
  ...props
}: MagicLinkVerifyViewProps) => {
  const { config } = useFrontier();

  const { mutateAsync: authCallback, isPending } = useMutation(
    FrontierServiceQueries.authCallback
  );
  const [emailParam, setEmailParam] = useState<string>('');
  const [stateParam, setStateParam] = useState<string>('');
  const [otp, setOTP] = useState<string>('');
  const [submitError, setSubmitError] = useState<string>('');
  const isButtonDisabledRef = useRef(true);

  const handleOTPChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;
    isButtonDisabledRef.current = value.length === 0;
    if (submitError.length > 0) setSubmitError('');
    setOTP(value);
  };

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const emailParam = params.get('email');
    const stateParam = params.get('state');

    emailParam && setEmailParam(emailParam);
    stateParam && setStateParam(stateParam);
  }, []);

  const OTPVerifyHandler = useCallback(
    async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      try {
        await authCallback({
          strategyName: 'mailotp',
          code: otp,
          state: stateParam
        });

        const destination = redirectURL ?? window.location.origin;
        window.location.replace(destination);
      } catch (error) {
        console.log(error);
        isButtonDisabledRef.current = true;
        setSubmitError('Please enter a valid OTP');
      }
    },
    [otp, stateParam, authCallback, redirectURL]
  );

  return (
    <AuthContainer {...props}>
      <Flex direction="column" gap={5}>
        <AuthHeader logo={logo} title={title} />
        {emailParam && (
          <Text size="small">
            We have sent an OTP. Please check your inbox at
            <b> {emailParam}</b>
          </Text>
        )}
      </Flex>

      <form onSubmit={OTPVerifyHandler} className={styles.form}>
        <Flex direction="column" gap={2} className={styles.otpInputContainer}>
          <InputField
            data-test-id="enter-code"
            autoFocus
            size="large"
            placeholder="Enter OTP"
            onChange={handleOTPChange}
            className={styles.textFieldCode}
          />

          <Text size="small" variant="danger" className={styles.error}>
            {submitError && String(submitError)}
          </Text>
        </Flex>

        <Button
          data-test-id="continue-with-login-code"
          className={styles.button}
          disabled={isButtonDisabledRef.current}
          type="submit"
          loading={isPending}
          loaderText="Submitting..."
        >
          Submit OTP
        </Button>
      </form>

      <Link href={config.redirectLogin || ''} data-test-id="back-to-login">
        <Text size="small">Back to login</Text>
      </Link>
    </AuthContainer>
  );
};
