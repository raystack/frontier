'use client';

import { Button, Flex, Link, Text, TextField } from '@raystack/apsara';
import React, {
  ComponentPropsWithRef,
  useCallback,
  useEffect,
  useState
} from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { Container } from './Container';
import { Header } from './Header';
import { hasWindow } from './helper';

const styles = {
  wrapper: {
    width: '80%'
  }
};

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

  const OTPVerifyClickHandler = useCallback(async () => {
    setLoading(true);
    try {
      await client.verifyMagicLinkAuthStrategyEndpoint(otp, stateParam!);

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
  }, [otp]);

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
          style={styles.wrapper}
          onClick={() => setVisiable(true)}
        >
          <Text>Enter code manually</Text>
        </Button>
      ) : (
        <Flex direction={'column'} style={styles.wrapper} gap="medium">
          <Flex direction="column">
            <TextField
              // @ts-ignore
              size="medium"
              key={'code'}
              placeholder="Enter code"
              onChange={handleOTPChange}
            />
            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {submitError && String(submitError)}
            </Text>
          </Flex>
          <Button
            size="medium"
            variant="primary"
            style={{ width: '100%' }}
            disabled={!otp}
            onClick={OTPVerifyClickHandler}
          >
            <Text style={{ color: 'var(--foreground-inverted)' }}>
              Continue with login code
            </Text>
          </Button>
        </Flex>
      )}
      <Link href={config.redirectLogin} style={{}}>
        <Text size={2}>Back to login</Text>
      </Link>
    </Container>
  );
};
