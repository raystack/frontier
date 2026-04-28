import { Button, Text } from '@raystack/apsara-v1';
import { HTMLProps } from 'react';
import GoogleLogo from '~/react/assets/logos/google-logo.svg';
import { capitalize } from '~/utils';
import styles from './oidc-button.module.css';

const oidcLogoMap = new Map([['google', GoogleLogo]]);

type OIDCButtonProps = HTMLProps<HTMLButtonElement> & {
  provider: string;
};

export const OIDCButton = ({ onClick, provider }: OIDCButtonProps) => (
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
