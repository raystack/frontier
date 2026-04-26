'use client';

import { useCallback, useMemo } from 'react';
import { Button, Skeleton, Text, Flex, toastManager } from '@raystack/apsara-v1';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { useFrontier } from '../../../contexts/FrontierContext';
import {
  Subscription,
  ChangeSubscriptionRequestSchema,
  FrontierServiceQueries
} from '@raystack/proton/frontier';
import { useMutation } from '@connectrpc/connect-query';
import { createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { create } from '@bufbuild/protobuf';
import dayjs from 'dayjs';
import {
  checkSimilarPlans,
  getPlanChangeAction,
  getPlanIntervalName,
  getPlanNameWithInterval
} from '../../../utils';
import { DEFAULT_DATE_FORMAT } from '../../../utils/constants';
import { timestampToDayjs } from '../../../../utils/timestamp';
import styles from '../billing-view.module.css';

interface UpcomingPlanChangeBannerProps {
  isLoading?: boolean;
  subscription?: Subscription;
  isAllowed: boolean;
}

export function UpcomingPlanChangeBanner({
  isLoading,
  subscription,
  isAllowed
}: UpcomingPlanChangeBannerProps) {
  const {
    config,
    activePlan,
    activeOrganization,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutate: cancelUpcomingChange, isPending: isPlanChangeLoading } =
    useMutation(FrontierServiceQueries.changeSubscription, {
      onSuccess: async () => {
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listSubscriptions,
            transport,
            input: {
              orgId: activeOrganization?.id ?? ''
            },
            cardinality: 'finite'
          })
        });
        toastManager.add({
          title: `Success: Your ${activePlan?.title} is resumed`,
          type: 'success'
        });
      },
      onError: (err: Error) => {
        toastManager.add({
          title: 'Failed to resume plan',
          description: err.message,
          type: 'error'
        });
      }
    });

  const phases =
    subscription?.phases?.filter(phase =>
      timestampToDayjs(phase.effectiveAt)?.isAfter(dayjs())
    ) || [];

  const nextPhase = phases?.[0];

  const upcomingPlan = useMemo(() => {
    if (nextPhase?.planId && allPlans.length > 0) {
      return allPlans.find(p => p.id === nextPhase?.planId);
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
    if (subscription?.id) {
      cancelUpcomingChange(
        create(ChangeSubscriptionRequestSchema, {
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
  }, [subscription?.id, cancelUpcomingChange]);

  const currentPlanName = getPlanNameWithInterval(activePlan);
  const upcomingPlanName = getPlanNameWithInterval(upcomingPlan || basePlan);

  const areSimilarPlans = checkSimilarPlans(activePlan, upcomingPlan);

  const resumePlanTitle = areSimilarPlans
    ? getPlanIntervalName(activePlan)
    : activePlan?.title;

  const showBanner =
    nextPhase?.planId ||
    (subscription?.state === 'active' && nextPhase?.reason === 'cancel');

  if (showLoader) return <Skeleton />;
  if (!showBanner) return null;

  return (
    <Flex className={styles.currentPlanInfoBox} justify="between" align="center">
      <Flex gap={3} align="center">
        <InfoCircledIcon />
        <Text size="small">
          Your {currentPlanName} will be{' '}
          {planAction?.btnDoneLabel.toLowerCase()} to {upcomingPlanName} from{' '}
          {expiryDate}.
        </Text>
      </Flex>
      {isAllowed ? (
        <Button
          variant="outline"
          color="neutral"
          size="small"
          onClick={onPlanChangeCancel}
          disabled={isPlanChangeLoading}
          loading={isPlanChangeLoading}
          loaderText="Loading..."
          data-test-id="frontier-sdk-upcoming-plan-change-banner-resume-button"
        >
          Resume with {resumePlanTitle}
        </Button>
      ) : null}
    </Flex>
  );
}
