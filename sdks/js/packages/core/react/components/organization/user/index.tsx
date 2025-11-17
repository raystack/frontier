'use client';

import { Flex } from '@raystack/apsara';
import { PageHeader } from '~/react/components/common/page-header';
import { UpdateProfile } from './update';
import sharedStyles from '../styles.module.css';

export function UserSetting() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <PageHeader 
            title="Profile" 
            description="Manage your profile information and settings."
          />
        </Flex>
        <UpdateProfile />
      </Flex>
    </Flex>
  );
}
