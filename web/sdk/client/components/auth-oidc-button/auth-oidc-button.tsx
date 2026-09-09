import { Button } from '@raystack/apsara';
import { ComponentProps } from 'react';
import GoogleLogo from '~/client/assets/logos/google-logo.svg';
import { capitalize } from '~/utils';
import styles from './auth-oidc-button.module.css';

const oidcLogoMap = new Map([['google', GoogleLogo]]);

export type AuthOIDCButtonProps = Omit<
  ComponentProps<typeof Button>,
  'variant' | 'color' | 'className' | 'leadingIcon' | 'children'
> & {
  provider: string;
};

export const AuthOIDCButton = ({
  onClick,
  provider,
  disabled,
  ...props
}: AuthOIDCButtonProps) => (
  <Button
    variant="outline"
    color="neutral"
    className={styles.button}
    onClick={onClick}
    disabled={disabled}
    data-test-id="frontier-sdk-oidc-logo-btn"
    {...props}
    leadingIcon={
      oidcLogoMap.has(provider) ? (
        <img
          src={oidcLogoMap.get(provider) as unknown as string}
          alt={provider + '-logo'}
        />
      ) : null
    }
  >
    Continue with {capitalize(provider)}
  </Button>
);
