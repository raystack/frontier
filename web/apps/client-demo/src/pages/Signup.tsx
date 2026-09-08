import useAuthRedirect from '@/hooks/useAuthRedirect';
import { useAuthRejection } from '@/hooks/useAuthRejection';
import { Flex } from '@raystack/apsara';
import { SignUpView } from '@raystack/frontier/client';

export default function Signup() {
  useAuthRedirect();

  // the callback page redirects here with the rejection it could not act on
  const rejection = useAuthRejection();

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignUpView error={rejection?.kind} errorStrategy={rejection?.strategy} />
    </Flex>
  );
}
