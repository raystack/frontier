import { Button, Flex, Text } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { V1Beta1Plan, V1Beta1Subscription } from '~/src';
import styles from './styles.module.css';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import dayjs from 'dayjs';
import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { getPlanChangeAction } from '~/react/utils';
import { toast } from 'sonner';

interface ChangeBannerProps {
  isLoading?: boolean;
  subscription?: V1Beta1Subscription;
}

export function UpcomingPlanChangeBanner({
  isLoading,
  subscription
}: ChangeBannerProps) {
  const {
    client,
    config,
    activePlan,
    activeOrganization,
    billingAccount,
    fetchActiveSubsciption
  } = useFrontier();
  const [upcomingPlan, setUpcomingPlan] = useState<V1Beta1Plan>();
  const [isPlanLoading, setIsPlanLoading] = useState(false);
  const [isPlanChangeLoading, setIsPlanChangeLoading] = useState(false);

  const nextPhase = subscription?.phases?.[0];

  const fetchPlan = useCallback(
    async (planId: string) => {
      setIsPlanLoading(true);
      try {
        const resp = await client?.frontierServiceGetPlan(planId);
        const plan = resp?.data?.plan;
        if (plan) {
          setUpcomingPlan(plan);
        } else {
          setUpcomingPlan(undefined);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setIsPlanLoading(false);
      }
    },
    [client]
  );

  useEffect(() => {
    if (nextPhase?.plan_id) {
      fetchPlan(nextPhase?.plan_id);
    }
  }, [fetchPlan, nextPhase?.plan_id]);

  const expiryDate = nextPhase?.effective_at
    ? dayjs(nextPhase?.effective_at).format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      )
    : '';

  const newPlanMetadata = upcomingPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const planAction = getPlanChangeAction(
    Number(newPlanMetadata?.weightage) || 0,
    Number(activePlanMetadata?.weightage)
  );

  const showLoader = isLoading || isPlanLoading;

  const onPlanChangeCancel = useCallback(async () => {
    setIsPlanChangeLoading(true);
    try {
      if (activeOrganization?.id && billingAccount?.id && subscription?.id) {
        const resp = await client?.frontierServiceChangeSubscription(
          activeOrganization?.id,
          billingAccount?.id,
          subscription?.id,
          {
            phase_change: {
              cancel_upcoming_changes: true
            }
          }
        );
        if (resp?.data?.phase) {
          await fetchActiveSubsciption();
          toast.success(`Success: Your ${activePlan?.title} is resumed`);
        }
      }
    } catch (err) {
      console.error(err);
    } finally {
      setIsPlanChangeLoading(false);
    }
  }, [
    activeOrganization?.id,
    activePlan?.title,
    billingAccount?.id,
    client,
    fetchActiveSubsciption,
    subscription?.id
  ]);

  return showLoader ? (
    <Skeleton />
  ) : nextPhase?.plan_id ? (
    <Flex className={styles.changeBannerBox} justify={'between'}>
      <Flex gap="small" className={styles.flex1} align={'center'}>
        <InfoCircledIcon className={styles.currentPlanInfoText} />
        <Text>
          Your {activePlan?.title} will be{' '}
          {planAction?.btnDoneLabel.toLowerCase()} to {upcomingPlan?.title} from{' '}
          {expiryDate}.
        </Text>
      </Flex>
      <Flex>
        <Button
          variant={'secondary'}
          onClick={onPlanChangeCancel}
          disabled={isPlanChangeLoading}
        >
          {isPlanChangeLoading
            ? 'Loading...'
            : `Resume with ${activePlan?.title}`}
        </Button>
      </Flex>
    </Flex>
  ) : null;
}
