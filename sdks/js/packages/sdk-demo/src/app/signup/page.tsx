'use client';
import { Flex } from '@raystack/apsara';
import { SignUp } from '@raystack/frontier/react';

export default function SignUpRoute() {
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignUp />
    </Flex>
  );
}
