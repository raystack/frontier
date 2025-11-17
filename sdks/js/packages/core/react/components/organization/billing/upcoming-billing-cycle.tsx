import { useNavigate } from '@tanstack/react-router';
import { ReactNode, useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
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
  Image,
  Skeleton,
  Text,
  Flex,
  Amount,
  toast
} from '@raystack/apsara';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import {
  getPlanIntervalName,
  getPlanNameWithInterval,
  makePlanSlug
} from '~/react/utils';
import { NEGATIVE_BALANCE_TOOLTIP_MESSAGE } from '~/react/utils/constants';
import { timestampToDayjs } from '~/utils/timestamp';
import line from '~/react/assets/line.svg';
import billingStyles from './billing.module.css';

function LabeledBillingData({
  label,
  value
}: {
  label: string;
  value: string | ReactNode;
}) {
  return (
    <Flex gap={2}>
      <Text size="small" weight="medium">
        {label}:
      </Text>
      <Text size="small">{value}</Text>
    </Flex>
  );
}

function PlanSwitchButton({ nextPlan }: { nextPlan: Plan }) {
  const intervalName = getPlanIntervalName(nextPlan).toLowerCase();

  const navigate = useNavigate({ from: '/billing' });
  function onClick() {
    navigate({
      to: '/billing/cycle-switch/$planId',
      params: {
        planId: nextPlan.id || ''
      }
    });
  }

  return (
    <div>
      <Button
        variant="text"
        color="accent"
        size="small"
        onClick={onClick}
        data-test-id="frontier-sdk-plan-switch-btn"
      >
        Switch to {intervalName}
      </Button>
    </div>
  );
}

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
}

export const UpcomingBillingCycle = ({
  isAllowed,
  isPermissionLoading
}: UpcomingBillingCycleProps) => {
  const {
    billingAccount,
    config,
    activeSubscription,
    trialSubscription,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();
  const navigate = useNavigate({ from: '/billing' });

  const {
    data: upcomingInvoice,
    isLoading: isInvoiceLoading,
    error: invoiceError
  } = useConnectQuery(
    FrontierServiceQueries.getUpcomingInvoice,
    create(GetUpcomingInvoiceRequestSchema, {
      orgId: billingAccount?.orgId || '',
      billingId: billingAccount?.id || ''
    }),
    {
      enabled:
        !!billingAccount?.id &&
        !!billingAccount?.orgId &&
        // This is to prevent fetching the upcoming invoice for offline billing accounts
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
          action: {
            label: 'Upgrade',
            link: '/plans'
          }
        }
      : {
          message: 'You are not subscribed to any plan',
          action: {
            label: 'Subscribe',
            link: '/plans'
          }
        };

  const onActionBtnClick = () => {
    // @ts-ignore
    navigate({ to: planInfo.action.link });
  };

  const alreadyPhased = activeSubscription?.phases?.find(
    phase => phase.planId === switchablePlan?.id
  );

  const error = memberCountError || invoiceError;
  useEffect(() => {
    if (error) {
      console.error('Failed to get upcoming billing cycle details', error);
      toast.error('Failed to get upcoming billing cycle details');
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

  return isLoading ? (
    <Skeleton />
  ) : dueDate && !isUserOnlyTrialing ? (
    <Flex
      align="center"
      justify="between"
      gap={3}
      className={billingStyles.billingCycleBox}
    >
      <Flex gap={5} align="center">
        <LabeledBillingData label="Plan" value={planName} />
        {switchablePlan && isAllowed && !alreadyPhased ? (
          <PlanSwitchButton
            nextPlan={switchablePlan}
            data-test-id="frontier-sdk-billing-cycle-interval-switch-button"
          />
        ) : null}
      </Flex>
      <Flex gap={5}>
        <LabeledBillingData
          label="Next billing"
          value={timestampToDayjs(dueDate)?.format(config.dateFormat) || '-'}
        />
        <Image src={line as unknown as string} alt="line" />
        <LabeledBillingData label="Users" value={memberCount} />
        <Image src={line as unknown as string} alt="line" />
        <LabeledBillingData
          label="Amount"
          value={
            <Flex gap={5}>
              <Amount
                currency={upcomingInvoice?.currency}
                value={Number(upcomingInvoice?.amount)}
              />
              {Number(upcomingInvoice?.amount) < 0 ? (
                <Tooltip
                  message={NEGATIVE_BALANCE_TOOLTIP_MESSAGE}
                  side="bottom"
                >
                  <InfoCircledIcon />
                </Tooltip>
              ) : null}
            </Flex>
          }
        />
      </Flex>
    </Flex>
  ) : (
    <Flex
      className={billingStyles.currentPlanInfoBox}
      align="center"
      justify="between"
      gap={3}
    >
      <Flex gap={3}>
        <InfoCircledIcon className={billingStyles.currentPlanInfoText} />
        <Text size="small" className={billingStyles.currentPlanInfoText}>
          {planInfo.message}
        </Text>
      </Flex>
      <Button
        variant="outline"
        color="neutral"
        size="small"
        onClick={onActionBtnClick}
        data-test-id="frontier-sdk-upcoming-billing-cycle-action-button"
      >
        {planInfo.action.label}
      </Button>
    </Flex>
  );
};
