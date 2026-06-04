import { Button, Text } from '@raystack/apsara';
import { HTMLProps } from 'react';
import GoogleLogo from '~/client/assets/logos/google-logo.svg';
import { capitalize } from '~/utils';
import styles from './auth-oidc-button.module.css';

const oidcLogoMap = new Map([['google', GoogleLogo]]);

export type AuthOIDCButtonProps = HTMLProps<HTMLButtonElement> & {
  provider: string;
};

export const AuthOIDCButton = ({ onClick, provider }: AuthOIDCButtonProps) => (
  <Button
    variant="outline"
    color="neutral"
    className={styles.button}
    onClick={onClick}
    data-test-id="frontier-sdk-oidc-logo-btn"
  >
    {oidcLogoMap.has(provider) ? (
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={oidcLogoMap.get(provider) as unknown as string}
        alt={provider + '-logo'}
        className={styles.logo}
      />
    ) : null}
    <Text size="regular">Continue with {capitalize(provider)}</Text>
  </Button>
);
