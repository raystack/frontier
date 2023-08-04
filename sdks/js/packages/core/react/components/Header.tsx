import { Flex, Text } from '@raystack/apsara';
import React, { ComponentPropsWithRef } from 'react';
import logo from '~/react/assets/logo.png';
import { useFrontier } from '../contexts/FrontierContext';

const styles = {
  container: {
    fontSize: '12px',
    minWidth: '220px',
    maxWidth: '100%',
    color: 'var(--foreground-base)',

    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '32px'
  },
  logoContainer: {},
  titleContainer: {
    fontWeight: '400'
  }
};

const defaultLogo = (
  // eslint-disable-next-line @next/next/no-img-element
  <img
    alt="logo"
    src={logo}
    style={{ borderRadius: 'var(--pd-8)', width: '80px', height: '80px' }}
  ></img>
);

type HeaderProps = ComponentPropsWithRef<'div'> & {
  title?: string;
  logo?: React.ReactNode;
};

export const Header = ({ title, logo }: HeaderProps) => {
  const { config } = useFrontier();

  return (
    <Flex
      style={{
        ...styles.container,
        flexDirection: 'column'
      }}
    >
      <div style={styles.logoContainer}>{logo ? logo : defaultLogo}</div>
      <div style={styles.titleContainer}>
        <Text size={9}>{title}</Text>
      </div>
    </Flex>
  );
};
