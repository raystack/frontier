'use client';

import { Button, Flex, Link, Text, TextField } from '@raystack/apsara';
import React, {
  ComponentPropsWithRef,
  useCallback,
  useEffect,
  useState
} from 'react';
import { Container } from '~/react/components/Container';
import { Header } from '~/react/components/Header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { hasWindow } from '~/utils/index';

// @ts-ignore
import styles from './onboarding.module.css';

type MagicLinkVerifyProps = ComponentPropsWithRef<typeof Container> & {
  logo?: React.ReactNode;
  title?: string;
};

export const MagicLinkVerify = ({
  logo,
  title = 'Check your email',
  ...props
}: MagicLinkVerifyProps) => {
  const [loading, setLoading] = useState<boolean>(false);
  const { client, config, strategies = [] } = useFrontier();
  const [visiable, setVisiable] = useState<boolean>(false);
  const [email, setEmail] = useState<string>('');
  const [emailParam, setEmailParam] = useState<string>('');
  const [stateParam, setStateParam] = useState<string>('');
  const [codeParam, setCodeParam] = useState<string>('');
  const [otp, setOTP] = useState<string>('');
  const [submitError, setSubmitError] = useState<string>('');

  const handleOTPChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setOTP(event.target.value);
  };

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const emailParam = params.get('email');
    const stateParam = params.get('state');
    const codeParam = params.get('code');

    emailParam && setEmailParam(emailParam);
    stateParam && setStateParam(stateParam);
    codeParam && setCodeParam(codeParam);
  }, []);

  const OTPVerifyHandler = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setLoading(true);
      try {
        if (!client) return;

        await client.frontierServiceAuthCallback({
          strategyName: 'mailotp',
          code: otp,
          state: stateParam
        });

        const searchParams = new URLSearchParams(
          hasWindow() ? window.location.search : ``
        );
        const redirectURL =
          searchParams.get('redirect_uri') || searchParams.get('redirectURL');

        // @ts-ignore
        window.location = redirectURL ? redirectURL : window.location.origin;
      } catch (error) {
        console.log(error);
        setSubmitError('Please enter a valid verification code');
      } finally {
        setLoading(false);
      }
    },
    [otp]
  );

  return (
    <Container {...props}>
      <Flex direction={'column'} gap="medium">
        <Header logo={logo} title={title} />
        {emailParam && (
          <Text>
            We have sent a temporary login link. Please check your inbox at
            <b> {emailParam}</b>
          </Text>
        )}
      </Flex>
      {!visiable ? (
        <Button
          variant="ghost"
          size="medium"
          className={styles.container80}
          onClick={() => setVisiable(true)}
        >
          <Text>Enter code manually</Text>
        </Button>
      ) : (
        <form
          onSubmit={OTPVerifyHandler}
          className={styles.container80}
          style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}
        >
          <Flex direction="column">
            <TextField
              // @ts-ignore
              size="medium"
              key={'code'}
              placeholder="Enter code"
              onChange={handleOTPChange}
            />
            <Text size={1} className={styles.error}>
              {submitError && String(submitError)}
            </Text>
          </Flex>
          <Button
            size="medium"
            variant="primary"
            className={styles.container}
            disabled={!otp}
            type="submit"
          >
            <Text className={styles.continue}>
              {loading ? 'Submitting...' : 'Continue with login code'}
            </Text>
          </Button>
        </form>
      )}
      <Link href={config.redirectLogin}>
        <Text size={2}>Back to login</Text>
      </Link>
    </Container>
  );
};
