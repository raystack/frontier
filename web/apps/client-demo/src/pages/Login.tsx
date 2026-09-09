import useAuthRedirect from '@/hooks/useAuthRedirect';
import { useAuthRejection } from '@/hooks/useAuthRejection';
import { Flex } from '@raystack/apsara';
import { SignInView } from '@raystack/frontier/client';

export default function Login() {
  useAuthRedirect();

  const rejection = useAuthRejection();

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignInView error={rejection} />
    </Flex>
  );
}
