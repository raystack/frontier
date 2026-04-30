import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara';
import { SignInView } from '@raystack/frontier/client';

export default function Login() {
  useAuthRedirect();
  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignInView />
    </Flex>
  );
}
