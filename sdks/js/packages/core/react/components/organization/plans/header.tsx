import { Flex, Skeleton, Link } from '@raystack/apsara';
import { PageHeader } from '~/react/components/common/page-header';
import plansStyles from './plans.module.css';

interface PlansHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

export const PlansHeader = ({ billingSupportEmail, isLoading }: PlansHeaderProps) => {
  if (isLoading) {
    return (
      <Flex direction="column" gap={2} className={plansStyles.flex1}>
        <Skeleton />
        <Skeleton />
      </Flex>
    );
  }

  return (
    <PageHeader
      title="Plans"
      description={
        billingSupportEmail ? (
          <>
            View and manage your subscription plan.{' '}
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
        ) : (
          'View and manage your subscription plan.'
        )
      }
    />
  );
};
