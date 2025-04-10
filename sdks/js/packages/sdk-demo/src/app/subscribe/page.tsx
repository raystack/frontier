'use client';
import { Flex } from '@raystack/apsara';
import { Subscribe } from '@raystack/frontier/react';
import React from 'react';

export default function SubscribeRoute() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <Subscribe  />
    </Flex>
  );
}
