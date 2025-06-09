'use client';

import { Flex } from '@raystack/apsara/v1';
import { MagicLinkVerify } from '@raystack/frontier/react';
import React from 'react';

export default function LoginRoute() {
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
