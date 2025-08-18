import { Text, Flex, Link } from '@raystack/apsara';

interface PlansHeaderProps {
  billingSupportEmail?: string;
}

export const PlansHeader = ({ billingSupportEmail }: PlansHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap={3}>
        <Text size="large">Plans</Text>
        <Text size="regular" variant="secondary">
          View and manage your subscription plan.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <Link
                size="regular"
                href={`mailto:${billingSupportEmail}`}
                data-test-id="frontier-sdk-billing-email-link"
                external
                style={{ textDecoration: 'none' }}
              >
                {billingSupportEmail}
              </Link>
            </>
          ) : null}
        </Text>
      </Flex>
    </Flex>
  );
};
