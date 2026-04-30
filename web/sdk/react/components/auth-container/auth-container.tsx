import { ComponentPropsWithRef, ReactNode } from 'react';
import { Flex } from '@raystack/apsara-v1';
import styles from './auth-container.module.css';

export type AuthContainerProps = ComponentPropsWithRef<'div'> & {
  children?: ReactNode;
  className?: string;
};

export const AuthContainer = ({
  children,
  style,
  className
}: AuthContainerProps) => {
  return (
    <Flex
      direction="column"
      align="center"
      gap={9}
      className={`${styles.container} ${className ?? ''}`}
      style={style}
    >
      {children}
    </Flex>
  );
};
