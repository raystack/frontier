import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara';
import { SignInView, type AuthRejection } from '@raystack/frontier/client';
import { useLocation } from 'react-router-dom';

export default function Login() {
  useAuthRedirect();

  const rejection = useLocation().state as AuthRejection | null;

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignInView error={rejection ?? undefined} />
    </Flex>
  );
}
