'use client';

import { useEffect, useMemo } from 'react';
import { useFrontier } from '../../../contexts/FrontierContext';
import {
  GetUpcomingInvoiceRequestSchema,
  FrontierServiceQueries,
  ListOrganizationUsersRequestSchema,
  Plan
} from '@raystack/proton/frontier';
import { useQuery as useConnectQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  Button,
  Tooltip,
  Skeleton,
  Text,
  Flex,
  Amount,
  toastManager
} from '@raystack/apsara-v1';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import {
  getPlanIntervalName,
  getPlanNameWithInterval,
  makePlanSlug
} from '../../../utils';
import { NEGATIVE_BALANCE_TOOLTIP_MESSAGE } from '../../../utils/constants';
import { timestampToDayjs } from '../../../../utils/timestamp';
import styles from '../billing-view.module.css';

function getSwitchablePlan(plans: Plan[], currentPlan: Plan) {
  const currentPlanMetaData =
    (currentPlan?.metadata as Record<string, string>) || {};
  const currentPlanSlug =
    currentPlanMetaData?.plan_group_id || makePlanSlug(currentPlan);
  const similarPlans = plans.filter(plan => {
    const metaData = (plan?.metadata as Record<string, string>) || {};
    const planSlug = metaData?.plan_group_id || makePlanSlug(plan);
    return currentPlanSlug === planSlug && plan.id !== currentPlan.id;
  });
  return similarPlans.length ? similarPlans[0] : null;
}

interface UpcomingBillingCycleProps {
  isAllowed: boolean;
  isPermissionLoading: boolean;
  onCycleSwitchClick?: (planId: string) => void;
  onNavigateToPlans?: () => void;
}

export function UpcomingBillingCycle({
  isAllowed,
  isPermissionLoading,
  onCycleSwitchClick,
  onNavigateToPlans
}: UpcomingBillingCycleProps) {
  const {
    billingAccount,
    config,
    activeSubscription,
    trialSubscription,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading,
    activeOrganization
  } = useFrontier();

  const {
    data: upcomingInvoice,
    isLoading: isInvoiceLoading,
    error: invoiceError
  } = useConnectQuery(
    FrontierServiceQueries.getUpcomingInvoice,
    create(GetUpcomingInvoiceRequestSchema, {
      orgId: activeOrganization?.id || ''
    }),
    {
      enabled:
        !!activeOrganization?.id &&
        !!billingAccount?.providerId,
      select: data => data?.invoice
    }
  );

  const {
    data: memberCount = 0,
    isLoading: isMemberCountLoading,
    error: memberCountError
  } = useConnectQuery(
    FrontierServiceQueries.listOrganizationUsers,
    create(ListOrganizationUsersRequestSchema, {
      id: billingAccount?.orgId || ''
    }),
    {
      enabled: !!billingAccount?.id && !!billingAccount?.orgId,
      select: data => data?.users?.length || 0
    }
  );

  const { plan, switchablePlan } = useMemo(() => {
    if (activeSubscription?.planId && allPlans.length > 0) {
      const currentPlan = allPlans.find(
        p => p.id === activeSubscription.planId
      );
      const otherPlan = currentPlan
        ? getSwitchablePlan(allPlans, currentPlan)
        : null;
      return { plan: currentPlan, switchablePlan: otherPlan };
    }
    return { plan: null, switchablePlan: null };
  }, [activeSubscription?.planId, allPlans]);

  const planName = activeSubscription
    ? getPlanNameWithInterval(plan ?? undefined)
    : getPlanNameWithInterval(basePlan);

  const planInfo =
    activeSubscription || basePlan
      ? {
          message: `You are subscribed to ${planName}.`,
          action: { label: 'Upgrade' }
        }
      : {
          message: 'You are not subscribed to any plan',
          action: { label: 'Subscribe' }
        };

  const alreadyPhased = activeSubscription?.phases?.find(
    phase => phase.planId === switchablePlan?.id
  );

  const error = memberCountError || invoiceError;
  useEffect(() => {
    if (error) {
      toastManager.add({
        title: 'Failed to get upcoming billing cycle details',
        type: 'error'
      });
    }
  }, [error]);

  const isLoading =
    isActiveOrganizationLoading ||
    isInvoiceLoading ||
    isMemberCountLoading ||
    isAllPlansLoading ||
    isPermissionLoading;

  const isUserOnlyTrialing = !activeSubscription?.id && trialSubscription?.id;
  const dueDate = upcomingInvoice?.dueDate || upcomingInvoice?.periodEndAt;

  if (isLoading) return <Skeleton />;

  if (dueDate && !isUserOnlyTrialing) {
    const switchableIntervalName = switchablePlan
      ? getPlanIntervalName(switchablePlan).toLowerCase()
      : '';

    return (
      <Flex
        align="center"
        justify="between"
        gap={3}
        className={styles.billingCycleBox}
      >
        <Flex gap={3} align="center">
          <Text size="small">
            <Text as="span" size="small" weight="medium">Plan:</Text>
            {' '}{planName}
          </Text>
          {switchablePlan && isAllowed && !alreadyPhased ? (
            <Button
              variant="text"
              color="accent"
              size="small"
              onClick={() => onCycleSwitchClick?.(switchablePlan.id || '')}
              data-test-id="frontier-sdk-plan-switch-btn"
            >
              Switch to {switchableIntervalName}
            </Button>
          ) : null}
        </Flex>
        <Flex gap={5} align="center">
          <Text size="small">
            <Text as="span" size="small" weight="medium">Next billing:</Text>
            {' '}{timestampToDayjs(dueDate)?.format(config.dateFormat) || '-'}
          </Text>
          <div className={styles.separator} />
          <Text size="small">
            <Text as="span" size="small" weight="medium">Users:</Text>
            {' '}{memberCount}
          </Text>
          <div className={styles.separator} />
          <Flex gap={2} align="center">
            <Text size="small">
              <Text as="span" size="small" weight="medium">Amount:</Text>
              {' '}
              <Amount
                currency={upcomingInvoice?.currency}
                value={Number(upcomingInvoice?.amount)}
              />
            </Text>
            {Number(upcomingInvoice?.amount) < 0 ? (
              <Tooltip>
                <Tooltip.Trigger render={<span />}>
                  <InfoCircledIcon />
                </Tooltip.Trigger>
                <Tooltip.Content>
                  {NEGATIVE_BALANCE_TOOLTIP_MESSAGE}
                </Tooltip.Content>
              </Tooltip>
            ) : null}
          </Flex>
        </Flex>
      </Flex>
    );
  }

  return (
    <Flex
      className={styles.currentPlanInfoBox}
      align="center"
      justify="between"
      gap={3}
    >
      <Flex gap={3} align="center">
        <InfoCircledIcon />
        <Text size="small">{planInfo.message}</Text>
      </Flex>
      <Button
        variant="outline"
        color="neutral"
        size="small"
        onClick={() => onNavigateToPlans?.()}
        data-test-id="frontier-sdk-upcoming-billing-cycle-action-button"
      >
        {planInfo.action.label}
      </Button>
    </Flex>
  );
}
