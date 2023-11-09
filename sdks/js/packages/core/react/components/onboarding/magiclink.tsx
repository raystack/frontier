import { Button, Flex, Separator, Text, TextField } from '@raystack/apsara';
import React, { useCallback, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { Controller, useForm } from 'react-hook-form';

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

const emailSchema = yup.object({
  email: yup.string().trim().email().required()
});

type FormData = yup.InferType<typeof emailSchema>;

export const MagicLink = ({ children, ...props }: MagicLinkProps) => {
  const { client, config } = useFrontier();
  const [visiable, setVisiable] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);

  const {
    watch,
    control,
    handleSubmit,
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
          callbackUrl: config.callbackUrl
        });

        const searchParams = new URLSearchParams({ state, email: data.email });

        // @ts-ignore
        window.location = `${
          config.redirectMagicLinkVerify
        }?${searchParams.toString()}`;
      } finally {
        setLoading(false);
      }
    },
    [client, config.callbackUrl, config.redirectMagicLinkVerify]
  );

  const email = watch('email', '');

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
    <form
      style={{ ...styles.container, flexDirection: 'column' }}
      onSubmit={handleSubmit(magicLinkHandler)}
    >
      <Separator />
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
        size="medium"
        variant="primary"
        {...props}
        style={{
          ...styles.button,
          ...(!email ? styles.disabled : {})
        }}
        disabled={!email}
        type="submit"
      >
        <Text style={{ color: 'var(--foreground-inverted)' }}>
          {loading ? 'loading...' : 'Continue with Email'}
        </Text>
      </Button>
    </form>
  );
};
