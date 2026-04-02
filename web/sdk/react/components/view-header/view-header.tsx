import { ComponentProps } from 'react';
import { Flex, Headline, Text } from '@raystack/apsara-v1';

export interface ViewHeaderProps extends ComponentProps<typeof Flex> {
  title: string;
  description?: string;
}

export function ViewHeader({ title, description, children, ...props }: ViewHeaderProps) {
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
