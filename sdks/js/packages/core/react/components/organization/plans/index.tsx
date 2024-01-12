import { Button, EmptyState, Flex, Text } from '@raystack/apsara';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect, useState } from 'react';
import { V1Beta1Plan } from '~/src';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import plansStyles from './plans.module.css';

const PlansLoader = () => {
  return (
    <Flex direction={'column'}>
      {[...new Array(15)].map(i => (
        <Skeleton containerClassName={plansStyles.flex1} key={`loader-${i}`} />
      ))}
    </Flex>
  );
};

const NoPlans = () => {
  return (
    <EmptyState style={{ marginTop: 160 }}>
      <Text size={5} style={{ fontWeight: 'bold' }}>
        No Plans Available
      </Text>
      <Text size={2}>
        Sorry, No plans available at this moment. Please try again later
      </Text>
    </EmptyState>
  );
};

interface PlansHeaderProps {
  billingSupportEmail?: string;
}

const PlansHeader = ({ billingSupportEmail }: PlansHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>Plans</Text>
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <a
                href={`mailto:${billingSupportEmail}`}
                target="_blank"
                style={{ fontWeight: 400, color: 'var(--foreground-accent)' }}
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

interface PlansListProps {
  plans: V1Beta1Plan[];
}

const PlansList = ({ plans = [] }: PlansListProps) => {
  if (plans.length === 0) return <NoPlans />;
  return <Flex>No Plans</Flex>;
};

export default function Plans() {
  const { config, client } = useFrontier();
  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);

  useEffect(() => {
    async function getPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceListPlans();
        if (resp?.data?.plans) {
          setPlans(resp?.data?.plans);
        }
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setIsPlansLoading(false);
      }
    }

    getPlans();
  }, [client]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Plans</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <PlansHeader billingSupportEmail={config.billing?.supportEmail} />
        </Flex>
        {isPlansLoading ? <PlansLoader /> : <PlansList plans={plans} />}
      </Flex>
    </Flex>
  );
}
