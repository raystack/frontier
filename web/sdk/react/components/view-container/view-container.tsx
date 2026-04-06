import { ComponentProps } from 'react';
import { Flex } from '@raystack/apsara-v1';
import styles from './view-container.module.css';
import { cx } from 'class-variance-authority';

export interface ViewContainerProps extends ComponentProps<typeof Flex> {
  contentProps?: Omit<ComponentProps<typeof Flex>, 'children'>;
}

export function ViewContainer({ children, contentProps, className, ...props }: ViewContainerProps) {
  return (
    <Flex direction="column" align="center" className={cx(styles.container, className)} {...props}>
      <Flex direction="column" gap={7} {...contentProps} className={cx(styles.content, contentProps?.className)}>
        {children}
      </Flex>
    </Flex>
  );
}
