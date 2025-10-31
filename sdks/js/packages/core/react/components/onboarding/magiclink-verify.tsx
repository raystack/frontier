'use client';

import { Button, Text, Link, Flex, InputField } from '@raystack/apsara';
import React, {
  ComponentPropsWithRef,
  useCallback,
  useEffect,
  useRef,
  useState
} from 'react';
import { Container } from '~/react/components/Container';
import { Header } from '~/react/components/Header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, FrontierServiceQueries } from '~hooks';

// @ts-ignore
import styles from './onboarding.module.css';

type MagicLinkVerifyProps = ComponentPropsWithRef<typeof Container> & {
  logo?: React.ReactNode;
  title?: string;
  redirectURL?: string;
};

export const MagicLinkVerify = ({
  logo,
  title = 'Check your email',
  redirectURL,
  ...props
}: MagicLinkVerifyProps) => {
  const { config } = useFrontier();

  const { mutateAsync: authCallback, isPending } = useMutation(
    FrontierServiceQueries.authCallback
  );
  const [emailParam, setEmailParam] = useState<string>('');
  const [stateParam, setStateParam] = useState<string>('');
  const [otp, setOTP] = useState<string>('');
  const [submitError, setSubmitError] = useState<string>('');
  const isButtonDisabledRef = useRef(true);

  const handleOTPChange = (event: React.ChangeEvent<HTMLInputElement>) => {
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
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      try {
        await authCallback({
          strategyName: 'mailotp',
          code: otp,
          state: stateParam
        });

        // @ts-ignore
        window.location = redirectURL ? redirectURL : window.location.origin;
      } catch (error) {
        console.log(error);
        isButtonDisabledRef.current = true;
        setSubmitError('Please enter a valid OTP');
      }
    },
    [otp, stateParam, authCallback, redirectURL]
  );

  return (
    <Container {...props}>
      <Flex direction="column" gap={5}>
        <Header logo={logo} title={title} />
        {emailParam && (
          <Text size="small">
            We have sent an OTP. Please check your inbox at
            <b> {emailParam}</b>
          </Text>
        )}
      </Flex>

      <form onSubmit={OTPVerifyHandler} className={styles.container80}>
        <Flex direction="column" gap={2} className={styles.optInputContainer}>
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
          className={styles.container}
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
    </Container>
  );
};
