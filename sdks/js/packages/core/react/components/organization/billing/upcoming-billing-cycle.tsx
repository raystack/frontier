import { useNavigate } from '@tanstack/react-router';
import { ReactNode, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Invoice } from '~/src';
import {
  Button,
  Tooltip,
  Image,
  toast,
  Skeleton,
  Text,
  Flex,
  Amount
} from '@raystack/apsara';
import dayjs from 'dayjs';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import {
  getPlanIntervalName,
  getPlanNameWithInterval,
  makePlanSlug
} from '~/react/utils';
import { NEGATIVE_BALANCE_TOOLTIP_MESSAGE } from '~/react/utils/constants';
import line from '~/react/assets/line.svg';
import billingStyles from './billing.module.css';
import { Plan } from '@raystack/proton/frontier';

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
  const [upcomingInvoice, setUpcomingInvoice] = useState<V1Beta1Invoice>();
  const {
    client,
    billingAccount,
    config,
    activeSubscription,
    trialSubscription,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();
  const [isInvoiceLoading, setIsInvoiceLoading] = useState(false);
  const [memberCount, setMemberCount] = useState(0);
  const [isMemberCountLoading, setIsMemberCountLoading] = useState(false);
  const navigate = useNavigate({ from: '/billing' });

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

  useEffect(() => {
    async function getMemberCount(orgId: string) {
      setIsMemberCountLoading(true);
      try {
        const resp = await client?.frontierServiceListOrganizationUsers(orgId);
        const count = resp?.data?.users?.length;
        if (count) {
          setMemberCount(count);
        }
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setIsMemberCountLoading(false);
      }
    }

    async function getUpcomingInvoice(orgId: string, billingId: string) {
      setIsInvoiceLoading(true);
      try {
        const resp = await client?.frontierServiceGetUpcomingInvoice(
          orgId,
          billingId
        );
        const invoice = resp?.data?.invoice;
        if (invoice && invoice.state) {
          setUpcomingInvoice(invoice);
        }
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setIsInvoiceLoading(false);
      }
    }

    if (
      billingAccount?.id &&
      billingAccount?.orgId &&
      billingAccount?.providerId
    ) {
      getUpcomingInvoice(billingAccount?.orgId, billingAccount?.id);
      getMemberCount(billingAccount?.orgId);
    }
  }, [
    client,
    billingAccount?.orgId,
    billingAccount?.id,
    billingAccount?.providerId
  ]);

  const planName = activeSubscription
    ? getPlanNameWithInterval(plan)
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

  const isLoading =
    isActiveOrganizationLoading ||
    isInvoiceLoading ||
    isMemberCountLoading ||
    isAllPlansLoading ||
    isPermissionLoading;

  const isUserOnlyTrialing = !activeSubscription?.id && trialSubscription?.id;
  const due_date = upcomingInvoice?.due_date || upcomingInvoice?.period_end_at;

  return isLoading ? (
    <Skeleton />
  ) : due_date && !isUserOnlyTrialing ? (
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
          value={dayjs(due_date).format(config.dateFormat)}
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
