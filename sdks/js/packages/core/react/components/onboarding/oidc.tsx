import { Button, Text } from '@raystack/apsara';
import React from 'react';
import GoogleLogo from '~/react/assets/logos/google-logo.svg';
import { capitalize } from '~/utils';

const oidcLogoMap = new Map([['google', GoogleLogo]]);

const styles = {
  button: {
    width: '100%'
  },
  logo: {
    marginRight: '4px'
  }
};

type ButtonProps = React.HTMLProps<HTMLButtonElement> & {
  provider: string;
};

export const OIDCButton = ({
  type = 'button',
  onClick,
  provider
}: ButtonProps) => (
  <Button
    size="medium"
    variant="secondary"
    style={styles.button}
    onClick={onClick}
  >
    {oidcLogoMap.has(provider) ? (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={oidcLogoMap.get(provider)}
        alt={provider + '-logo'}
        style={styles.logo}
      />
    ) : null}
    <Text>Continue with {capitalize(provider)}</Text>
  </Button>
);
