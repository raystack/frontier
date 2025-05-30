'use client';
import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara/v1';
import { SignUp } from '@raystack/frontier/react';

export default function SignUpRoute() {
  useAuthRedirect();

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
