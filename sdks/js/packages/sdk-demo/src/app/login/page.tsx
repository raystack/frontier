'use client';
import { Flex } from '@raystack/apsara';
import { SignIn } from '@raystack/frontier/react';

export default function LoginRoute() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignIn />
    </Flex>
  );
}
