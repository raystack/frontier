'use client';

import { Flex } from '@raystack/apsara';
import { styles as sharedStyles } from '../styles';
import { PageHeader } from '~/react/components/common/page-header';
import { UpdateProfile } from './update';

export function UserSetting() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" style={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" style={sharedStyles.header}>
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
