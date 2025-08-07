import { Flex } from '@raystack/apsara';
import { MagicLinkVerify } from '@raystack/frontier/react';
import React from 'react';

export default function MagiclinkVerify() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <MagicLinkVerify />
    </Flex>
  );
}
