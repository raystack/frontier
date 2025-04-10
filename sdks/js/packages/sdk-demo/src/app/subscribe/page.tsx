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
      style={{ height: '95vh', width: '100vw' }}
    >
      <Subscribe
        title="Some title"
        desc="Some description with long text to see how it looks"
        onSubmit={data => console.log(JSON.stringify(data))}
        successTitle="Some confirm title"
        successDesc="Some confirm description with long text to see how it looks"
        activity="SomeActivity"
        source="SomeSource"
        medium="SomeMedium"
      />
    </Flex>
  );
}
