import { Flex, Text } from '@raystack/apsara';
import React, { ComponentPropsWithRef } from 'react';
import logo from '~/react/assets/logo.png';

// @ts-ignore
import styles from './header.module.css';

const defaultLogo = (
  // eslint-disable-next-line @next/next/no-img-element
  <img
    alt="logo"
    src={logo}
    style={{ borderRadius: 'var(--rs-space-3)', width: '80px', height: '80px' }}
  ></img>
);

type HeaderProps = ComponentPropsWithRef<'div'> & {
  title?: string;
  logo?: React.ReactNode;
};

export const Header = ({ title, logo }: HeaderProps) => {
  return (
    <Flex
      direction="column"
      className={styles.container}
      align="center"
      gap="large"
    >
      <div>{logo ? logo : defaultLogo}</div>
      <div className={styles.title}>
        <Text size={9}>{title}</Text>
      </div>
    </Flex>
  );
};
