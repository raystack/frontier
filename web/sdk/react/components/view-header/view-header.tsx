import { ComponentProps, ReactNode } from 'react';
import { Flex, Headline, Text } from '@raystack/apsara-v1';

export interface ViewHeaderProps extends ComponentProps<typeof Flex> {
  title: string;
  description?: string;
  /**
   * When provided, renders a breadcrumb row above the title.
   * `children` are placed alongside the breadcrumb (e.g. action buttons).
   * Without breadcrumb, `children` render alongside the title (default layout).
   */
  breadcrumb?: ReactNode;
}

export function ViewHeader({ title, description, breadcrumb, children, ...props }: ViewHeaderProps) {
  if (breadcrumb) {
    return (
      <Flex direction="column" gap={7} {...props}>
        <Flex direction="row" align="center" gap={5}>
          {breadcrumb}
          {children}
        </Flex>
        <Headline size="t1" weight="medium">
          {title}
        </Headline>
      </Flex>
    );
  }

  return (
    <Flex direction="row" justify="between" align="center" {...props}>
      <Flex direction="column" gap={3}>
        <Headline size="t1" weight="medium">
          {title}
        </Headline>
        {description ? (
          <Text size="regular" variant="secondary">
            {description}
          </Text>
        ) : null}
      </Flex>
      {children}
    </Flex>
  );
}
