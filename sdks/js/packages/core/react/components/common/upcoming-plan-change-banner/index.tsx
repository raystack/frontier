import { Button, Skeleton, Text, Flex, toast } from '@raystack/apsara';
import {
  DEFAULT_DATE_FORMAT,
  SUBSCRIPTION_STATES
} from '~/react/utils/constants';
import { Subscription, ChangeSubscriptionRequestSchema } from '@raystack/proton/frontier';
import styles from './styles.module.css';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import dayjs from 'dayjs';
import { useCallback, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, FrontierServiceQueries } from '~hooks';
import { create } from '@bufbuild/protobuf';
import { useQueryClient } from '@tanstack/react-query';
import { createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import {
  checkSimilarPlans,
  getPlanChangeAction,
  getPlanIntervalName,
  getPlanNameWithInterval
} from '~/react/utils';
import { timestampToDayjs } from '~/utils/timestamp';

interface ChangeBannerProps {
  isLoading?: boolean;
  subscription?: Subscription;
  isAllowed: boolean;
}

export function UpcomingPlanChangeBanner({
  isLoading,
  subscription,
  isAllowed
}: ChangeBannerProps) {
  const {
    config,
    activePlan,
    activeOrganization,
    billingAccount,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutate: cancelUpcomingChange, isPending: isPlanChangeLoading } = useMutation(
    FrontierServiceQueries.changeSubscription,
    {
      onSuccess: async () => {
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listSubscriptions,
            transport,
            input: {
              orgId: activeOrganization?.id ?? '',
              billingId: billingAccount?.id ?? ''
            },
            cardinality: 'finite'
          })
        });
        toast.success(`Success: Your ${activePlan?.title} is resumed`);
      },
      onError: (err) => {
        console.error(err);
        toast.error('Failed to resume plan', {
          description: err.message
        });
      }
    }
  );

  const phases =
    subscription?.phases?.filter(phase =>
      timestampToDayjs(phase.effectiveAt)?.isAfter(dayjs())
    ) || [];

  const nextPhase = phases?.[0];

  const upcomingPlan = useMemo(() => {
    if (nextPhase?.planId && allPlans.length > 0) {
      const plan = allPlans.find(p => p.id === nextPhase?.planId);
      return plan;
    }
  }, [nextPhase?.planId, allPlans]);

  const expiryDate = nextPhase?.effectiveAt
    ? timestampToDayjs(nextPhase?.effectiveAt)?.format(
        config?.dateFormat || DEFAULT_DATE_FORMAT
      )
    : '';

  const newPlanMetadata = upcomingPlan?.metadata as Record<string, number>;
  const activePlanMetadata = activePlan?.metadata as Record<string, number>;

  const planAction = getPlanChangeAction(
    Number(newPlanMetadata?.weightage),
    Number(activePlanMetadata?.weightage)
  );

  const showLoader = isLoading || isAllPlansLoading;

  const onPlanChangeCancel = useCallback(() => {
    if (activeOrganization?.id && billingAccount?.id && subscription?.id) {
      cancelUpcomingChange(
        create(ChangeSubscriptionRequestSchema, {
          orgId: activeOrganization?.id,
          billingId: billingAccount?.id,
          id: subscription?.id,
          change: {
            case: 'phaseChange',
            value: {
              cancelUpcomingChanges: true
            }
          }
        })
      );
    }
  }, [
    activeOrganization?.id,
    billingAccount?.id,
    subscription?.id,
    cancelUpcomingChange
  ]);

  const currentPlanName = getPlanNameWithInterval(activePlan);
  const upcomingPlanName = getPlanNameWithInterval(upcomingPlan || basePlan);

  const areSimilarPlans = checkSimilarPlans(activePlan, upcomingPlan);

  const resumePlanTitle = areSimilarPlans
    ? getPlanIntervalName(activePlan)
    : activePlan?.title;

  const showBanner =
    nextPhase?.planId ||
    (subscription?.state === SUBSCRIPTION_STATES.ACTIVE &&
      nextPhase?.reason === 'cancel');

  return showLoader ? (
    <Skeleton />
  ) : showBanner ? (
    <Flex className={styles.changeBannerBox} justify="between">
      <Flex gap={3} className={styles.flex1} align="center">
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
            variant="outline"
            color="neutral"
            size="small"
            onClick={onPlanChangeCancel}
            disabled={isPlanChangeLoading}
            loading={isPlanChangeLoading}
            loaderText="Loading..."
          >
            Resume with {resumePlanTitle}
          </Button>
        ) : null}
      </Flex>
    </Flex>
  ) : null;
}
