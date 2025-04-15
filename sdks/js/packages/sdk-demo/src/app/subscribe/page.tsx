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
        title="Some title"
        desc="Some description with long text to see how it looks"
        onSubmit={data => console.log(JSON.stringify(data))}
        activity="SomeActivity"
        source="SomeSource"
        medium="SomeMedium"
      />
    </Flex>
  );
}
