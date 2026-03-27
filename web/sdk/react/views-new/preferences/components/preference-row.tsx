import { ComponentProps } from 'react';
import { Flex, Text, Skeleton } from '@raystack/apsara-v1';
import styles from './preferences-row.module.css';
import { cx } from 'class-variance-authority';

export interface PreferenceRowProps extends ComponentProps<typeof Flex> {
  title: string;
  description: string;
  isLoading?: boolean;
  contentProps?: ComponentProps<typeof Flex>;
}

export function PreferenceRow({
  title,
  description,
  isLoading,
  children,
  className,
  contentProps,
  ...props
}: PreferenceRowProps) {
  return (
    <Flex align="center" gap={9} className={cx(styles.row, className)} {...props}>
      <Flex direction="column" gap={3} className={styles.content}>
        {isLoading ? (
          <>
            <Skeleton width="20%" height={24} />
            <Skeleton width="40%" height={16} />
          </>
        ) : (
          <>
            <Text size="large" weight="medium">
              {title}
            </Text>
            <Text size="small" variant="secondary">
              {description}
            </Text>
          </>
        )}
      </Flex>
      <Flex {...contentProps} className={cx(styles.children, contentProps?.className)}>
        {isLoading ? <Skeleton width={135} height={33} /> : children}
      </Flex>
    </Flex>
  );
}
