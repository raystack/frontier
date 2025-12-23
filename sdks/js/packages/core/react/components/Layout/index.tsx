import { Flex } from '@raystack/apsara';
import { PropsWithChildren } from 'react';
import { PageHeader } from '../common/page-header';
import sharedStyles from '../organization/styles.module.css';

interface LayoutProps {
  title: string;
  description?: string;
}

export function Layout({ title, description, children }: PropsWithChildren<LayoutProps>) {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <PageHeader title={title} description={description} />
        </Flex>
        <Flex direction="column" gap={9}>
          {children}
        </Flex>
      </Flex>
    </Flex>
  );
}
