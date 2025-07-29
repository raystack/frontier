import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara/v1';
import { SignIn } from '@raystack/frontier/react';

export default function Login() {
  useAuthRedirect();
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