'use client';
import { Flex } from '@raystack/apsara/v1';
import { Subscribe } from '@raystack/frontier/react';
import React from 'react';
export default function SubscribeRoute() {
  return (
    <Flex
      justify="center"
      direction="column"
      align="center"
      style={{ width: '100vw', height: '95vh' }}
    >
      <Subscribe
        onSubmit={data => console.log(JSON.stringify(data))}
      />
    </Flex>
  );
}
