import { Button, Text } from '@raystack/apsara';
import React from 'react';
import { capitalize } from './helper';
import GoogleLogo from '../assets/logos/google-logo.svg';

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
      <img
        src={oidcLogoMap.get(provider)}
        alt={provider + '-logo'}
        style={styles.logo}
      />
    ) : null}
    <Text>Continue with {capitalize(provider)}</Text>
  </Button>
);
