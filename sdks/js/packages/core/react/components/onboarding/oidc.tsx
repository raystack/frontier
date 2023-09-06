import { Button, Text } from '@raystack/apsara';
import React from 'react';
import GoogleLogo from '~/react/assets/logos/google-logo.svg';
import { capitalize } from '~/utils';
// @ts-ignore
import styles from './onboarding.module.css';

const oidcLogoMap = new Map([['google', GoogleLogo]]);
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
    className={styles.container}
    onClick={onClick}
  >
    {oidcLogoMap.has(provider) ? (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={oidcLogoMap.get(provider)}
        alt={provider + '-logo'}
        style={{ marginRight: '4px' }}
      />
    ) : null}
    <Text>Continue with {capitalize(provider)}</Text>
  </Button>
);
