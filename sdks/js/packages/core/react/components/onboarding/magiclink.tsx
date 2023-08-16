import { Button, Separator, Text, TextField } from '@raystack/apsara';
import React, { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';

const styles = {
  container: {
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    gap: 'var(--pd-16)'
  },

  button: {
    width: '100%'
  },
  disabled: { opacity: 1 }
};

type MagicLinkProps = {
  children?: React.ReactNode;
};
export const MagicLink = ({ children, ...props }: MagicLinkProps) => {
  const { client, config } = useFrontier();
  const [visiable, setVisiable] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const [email, setEmail] = useState<string>('');
  const [state, setState] = useState<string>('');

  const magicLinkClickHandler = useCallback(async () => {
    setLoading(true);
    try {
      if (!client) return;

      const {
        data: { state = '' }
      } = await client.frontierServiceAuthenticate('mailotp', {
        email,
        callbackUrl: config.callbackUrl
      });

      const searchParams = new URLSearchParams({ state, email });

      // @ts-ignore
      window.location = `${
        config.redirectMagicLinkVerify
      }?${searchParams.toString()}`;
    } finally {
      setLoading(false);
    }
  }, [client, config.callbackUrl, config.redirectMagicLinkVerify, email]);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  };

  if (!visiable)
    return (
      <Button
        variant="secondary"
        size="medium"
        style={styles.button}
        onClick={() => setVisiable(true)}
      >
        <Text>Continue with Email</Text>
      </Button>
    );

  return (
    <div style={{ ...styles.container, flexDirection: 'column' }}>
      <Separator />
      <TextField
        // @ts-ignore
        size="medium"
        key={'email'}
        placeholder="name@example.com"
        onChange={handleChange}
      />

      <Button
        size="medium"
        variant="primary"
        {...props}
        style={{
          ...styles.button,
          ...(!email ? styles.disabled : {})
        }}
        disabled={!email}
        onClick={magicLinkClickHandler}
      >
        <Text style={{ color: 'var(--foreground-inverted)' }}>
          {loading ? 'loading...' : 'Continue with Emails'}
        </Text>
      </Button>
    </div>
  );
};
