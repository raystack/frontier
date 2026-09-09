import useAuthRedirect from '@/hooks/useAuthRedirect';
import { Flex } from '@raystack/apsara';
import { SignUpView, type AuthRejection } from '@raystack/frontier/client';
import { useLocation } from 'react-router-dom';

export default function Signup() {
  useAuthRedirect();

  const rejection = useLocation().state as AuthRejection | null;

  return (
    <Flex
      justify="center"
      align="center"
      style={{ height: '100vh', width: '100vw' }}
    >
      <SignUpView error={rejection ?? undefined} />
    </Flex>
  );
}
