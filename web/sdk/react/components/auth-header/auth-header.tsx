import { ComponentPropsWithRef, ReactNode } from 'react';
import { Flex, Headline } from '@raystack/apsara-v1';
import logo from '~/react/assets/logo.png';
import styles from './auth-header.module.css';

const defaultLogo = (
  // eslint-disable-next-line @next/next/no-img-element
  <img alt="logo" src={logo} className={styles.logo} />
);

export type AuthHeaderProps = ComponentPropsWithRef<'div'> & {
  title?: string;
  logo?: ReactNode;
};

export const AuthHeader = ({ title, logo }: AuthHeaderProps) => {
  return (
    <Flex
      direction="column"
      className={styles.container}
      align="center"
      gap={9}
    >
      <div>{logo ? logo : defaultLogo}</div>
      <div className={styles.title}>
        <Headline size="t2">{title}</Headline>
      </div>
    </Flex>
  );
};
