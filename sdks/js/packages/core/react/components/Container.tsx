import { Flex } from '@raystack/apsara';
import { cva } from 'class-variance-authority';
import React, { ComponentPropsWithRef } from 'react';
// @ts-ignore
import styles from './container.module.css';

type ContainerProps = ComponentPropsWithRef<'div'> & {
  children?: React.ReactNode;
  shadow?: 'none' | 'xs' | 'sm' | 'md' | 'lg';
  radius?: 'none' | 'xs' | 'sm' | 'md' | 'lg';
  className?: string;
};

const container = cva(styles.container, {
  variants: {
    shadow: {
      none: '',
      xs: styles.shadowxs,
      sm: styles.shadowsm,
      md: styles.shadowmd,
      lg: styles.shadowlg
    },
    radius: {
      none: '',
      xs: styles.radiusxs,
      sm: styles.radiussm,
      md: styles.radiusmd,
      lg: styles.radiuslg
    }
  },
  defaultVariants: {
    shadow: 'none',
    radius: 'none'
  }
});

export const Container = ({
  children,
  shadow = 'none',
  radius = 'md',
  style,
  className
}: ContainerProps) => {
  return (
    <Flex
      direction="column"
      align="center"
      gap="large"
      className={`${container({ shadow, radius, className })} ${className}`}
      style={style}
    >
      {children}
    </Flex>
  );
};
