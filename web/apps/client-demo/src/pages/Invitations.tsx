import AuthContext from '@/contexts/auth';
import { Button, Flex, Text, toast } from '@raystack/apsara';
import {
  useQuery,
  useMutation,
  create,
  FrontierServiceQueries,
  useQueryClient,
} from '@raystack/frontier/hooks';
import { useContext, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

export default function Invitations() {
  const { isAuthorized } = useContext(AuthContext);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const {
    data: invitationsData,
    isLoading,
    error,
  } = useQuery(FrontierServiceQueries.listCurrentUserInvitations, {}, {
    enabled: isAuthorized,
  });

  const { mutateAsync: acceptInvitation } = useMutation(
    FrontierServiceQueries.acceptOrganizationInvitation,
  );

  const invitations = invitationsData?.invitations ?? [];
  const orgs = invitationsData?.orgs ?? [];

  const orgMap = orgs.reduce(
    (acc, org) => {
      acc[org.id] = org.title || org.name || org.id;
      return acc;
    },
    {} as Record<string, string>,
  );

  const handleAccept = useCallback(
    async (invitation: { id: string; orgId: string }) => {
      try {
        await acceptInvitation({ id: invitation.id, orgId: invitation.orgId });
        toast.success('Invitation accepted');
        queryClient.invalidateQueries();
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Something went wrong';
        toast.error(`Failed to accept invitation: ${message}`);
      }
    },
    [acceptInvitation, queryClient],
  );

  if (!isAuthorized) {
    navigate('/login');
    return null;
  }

  return (
    <main>
      <Flex
        direction="column"
        gap="medium"
        style={{ maxWidth: 600, margin: '0 auto', padding: 'var(--rs-space-6)' }}
      >
        <Flex justify="between" align="center">
          <Text size="large" weight="bold">
            Pending Invitations
          </Text>
          <Button
            variant="outline"
            color="neutral"
            size="small"
            onClick={() => navigate('/')}
          >
            Back
          </Button>
        </Flex>

        {isLoading ? (
          <Text variant="secondary">Loading invitations...</Text>
        ) : error ? (
          <Text variant="secondary">Failed to load invitations.</Text>
        ) : invitations.length === 0 ? (
          <Text variant="secondary">No pending invitations.</Text>
        ) : (
          <Flex direction="column" gap="small">
            {invitations.map((invitation) => (
              <Flex
                key={invitation.id}
                justify="between"
                align="center"
                style={{
                  padding: 'var(--rs-space-4)',
                  border: '1px solid var(--rs-color-border-base-secondary)',
                  borderRadius: 'var(--rs-radius-2)',
                }}
              >
                <Flex direction="column" gap="extra-small">
                  <Text weight="medium">
                    {orgMap[invitation.orgId] || invitation.orgId}
                  </Text>
                  <Text size="small" variant="secondary">
                    Invited as {invitation.userId}
                  </Text>
                </Flex>
                <Button
                  size="small"
                  data-test-id={`[accept-invite-${invitation.id}]`}
                  onClick={() => handleAccept(invitation)}
                >
                  Accept
                </Button>
              </Flex>
            ))}
          </Flex>
        )}
      </Flex>
    </main>
  );
}
