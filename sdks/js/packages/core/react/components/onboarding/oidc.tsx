import { Button, Text } from '@raystack/apsara/v1';
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
  onClick,
  provider
}: ButtonProps) => (
  <Button
    variant="outline"
    color="neutral"
    className={styles.container}
    onClick={onClick}
    data-test-id="frontier-sdk-oidc-logo-btn"
  >
    {oidcLogoMap.has(provider) ? (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={oidcLogoMap.get(provider) as unknown as string}
        alt={provider + '-logo'}
        style={{ marginRight: '4px' }}
      />
    ) : null}
    <Text size="regular">Continue with {capitalize(provider)}</Text>
  </Button>
);
