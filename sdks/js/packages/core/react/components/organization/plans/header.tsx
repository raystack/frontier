import { Flex } from '@raystack/apsara';
import { Text } from '@raystack/apsara/v1';

interface PlansHeaderProps {
  billingSupportEmail?: string;
}

export const PlansHeader = ({ billingSupportEmail }: PlansHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size="large">Plans</Text>
        <Text size="regular" variant="secondary">
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <a
                href={`mailto:${billingSupportEmail}`}
                data-test-id="frontier-sdk-billing-email-link"
                target="_blank"
                style={{ fontWeight: 'var(--rs-font-weight-regular)', color: 'var(--rs-color-foreground-accent-primary)' }}
              >
                {billingSupportEmail}
              </a>
            </>
          ) : null}
        </Text>
      </Flex>
    </Flex>
  );
};
