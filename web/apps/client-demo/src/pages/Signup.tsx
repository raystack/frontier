import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara';
import { SignUpView } from '@raystack/frontier/client';

export default function Signup() {
  useAuthRedirect();

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignUpView />
    </Flex>
  );
}
