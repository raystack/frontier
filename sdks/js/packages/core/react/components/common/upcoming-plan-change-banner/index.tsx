import { Button, Flex, Text } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';
import {
  DEFAULT_DATE_FORMAT,
  SUBSCRIPTION_STATES
} from '~/react/utils/constants';
import { V1Beta1Plan, V1Beta1Subscription } from '~/src';
import styles from './styles.module.css';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import dayjs from 'dayjs';
import { useCallback, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  checkSimilarPlans,
  getPlanChangeAction,
  getPlanIntervalName,
  getPlanNameWithInterval
} from '~/react/utils';
import { toast } from 'sonner';

interface ChangeBannerProps {
  isLoading?: boolean;
  subscription?: V1Beta1Subscription;
  isAllowed: boolean;
}

export function UpcomingPlanChangeBanner({
  isLoading,
  subscription,
  isAllowed
}: ChangeBannerProps) {
  const {
    client,
    config,
    activePlan,
    activeOrganization,
    billingAccount,
    fetchActiveSubsciption,
    basePlan
  } = useFrontier();
  const [upcomingPlan, setUpcomingPlan] = useState<V1Beta1Plan>();
  const [isPlanLoading, setIsPlanLoading] = useState(false);
  const [isPlanChangeLoading, setIsPlanChangeLoading] = useState(false);

  const phases =
    subscription?.phases?.filter(phase =>
      dayjs(phase.effective_at).isAfter(dayjs())
    ) || [];

  const nextPhase = phases?.[0];

  const fetchPlan = useCallback(
    async (planId: string) => {
      setIsPlanLoading(true);
      try {
        // This API call can be moved to sdks/js/packages/core/react/components/organization/billing/index.tsx
        const resp = await client?.frontierServiceGetPlan(planId);
        const plan = resp?.data?.plan ?? {};
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
    Number(newPlanMetadata?.weightage),
    Number(activePlanMetadata?.weightage)
  );

  const showLoader = isLoading || isPlanLoading;

  const onPlanChangeCancel = useCallback(async () => {
    setIsPlanChangeLoading(true);
    try {
      if (activeOrganization?.id && billingAccount?.id && subscription?.id) {
        // This API call can be moved to sdks/js/packages/core/react/components/organization/billing/index.tsx
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

  const currentPlanName = getPlanNameWithInterval(activePlan);
  const upcomingPlanName = getPlanNameWithInterval(upcomingPlan || basePlan);

  const areSimilarPlans = checkSimilarPlans(
    activePlan || {},
    upcomingPlan || {}
  );

  const resumePlanTitle = areSimilarPlans
    ? getPlanIntervalName(activePlan)
    : activePlan?.title;

  const showBanner =
    nextPhase?.plan_id ||
    (subscription?.state === SUBSCRIPTION_STATES.ACTIVE &&
      nextPhase?.reason === 'cancel');

  return showLoader ? (
    <Skeleton />
  ) : showBanner ? (
    <Flex className={styles.changeBannerBox} justify={'between'}>
      <Flex gap="small" className={styles.flex1} align={'center'}>
        <InfoCircledIcon className={styles.currentPlanInfoText} />
        <Text>
          Your {currentPlanName} will be{' '}
          {planAction?.btnDoneLabel.toLowerCase()} to {upcomingPlanName} from{' '}
          {expiryDate}.
        </Text>
      </Flex>
      <Flex>
        {isAllowed ? (
          <Button
            data-test-id="frontier-sdk-upcoming-plan-change-banner-resume-button"
            variant={'secondary'}
            onClick={onPlanChangeCancel}
            disabled={isPlanChangeLoading}
          >
            {isPlanChangeLoading
              ? 'Loading...'
              : `Resume with ${resumePlanTitle}`}
          </Button>
        ) : null}
      </Flex>
    </Flex>
  ) : null;
}
