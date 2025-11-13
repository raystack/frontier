'use client';

import { Flex, Headline, Text } from '@raystack/apsara';
import { ReactNode } from 'react';

interface PageHeaderProps {
  title: string;
  description?: string | ReactNode;
}

/**
 * PageHeader - Component for displaying page header with a title and optional description on top of SDK pages.
 * 
 * @param title - The main heading text for the page
 * @param description - Optional description text or ReactNode displayed below the title
 * 
 * @example
 * <PageHeader 
 *   title="Sessions" 
 *   description="Devices logged into this account."
 * />
 */
export const PageHeader = ({ title, description }: PageHeaderProps) => {
  return (
    <Flex direction="column" gap={2}>
      <Headline size="t1">{title}</Headline>
      {description && (
        <Text size="regular" variant="secondary">
          {description}
        </Text>
      )}
    </Flex>
  );
};

