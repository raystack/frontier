import AuthContext from '@/contexts/auth';
import { Button, Flex } from '@raystack/apsara';
import { useFrontier } from '@raystack/frontier/react';
import { useMutation, FrontierServiceQueries } from '@raystack/frontier/hooks';
import { Link, useNavigate } from 'react-router-dom';
import { useContext, useEffect } from 'react';

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);
  const { organizations } = useFrontier();
  const navigate = useNavigate();

  const logoutMutation = useMutation(FrontierServiceQueries.authLogout);

  useEffect(() => {
    if (!isAuthorized) {
      navigate('/login');
    }
  }, [isAuthorized, navigate]);

  async function logout() {
    try {
      await logoutMutation.mutateAsync({});
      window.location.reload();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }

  return (
    <main>
      <Flex
        align="center"
        style={{ height: '100vh', width: '100vw' }}
        direction="column"
      >
        <Button
          variant="outline"
          color="neutral"
          size="small"
          data-test-id="[logout-button]"
          onClick={logout}
        >
          Logout
        </Button>
        <Flex direction="row" wrap="wrap" gap={'medium'}>
          {organizations.map(org => (
            <Flex
              key={org.id}
              style={{
                padding: 'var(--rs-space-5)',
                border: '1px solid var(--rs-color-border-base-secondary)'
              }}
            >
              <Link
                to={`/organizations/${org.id}`}
                data-test-id={`[organization-link-${org.id}]`}
              >
                {org.title}
              </Link>
            </Flex>
          ))}
        </Flex>
      </Flex>
    </main>
  );
}
