import { useNavigate } from '@tanstack/react-router';
import { ReactNode, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Invoice, V1Beta1Plan } from '~/src';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import { Flex, Text, Image, Button, Tooltip } from '@raystack/apsara';
import billingStyles from './billing.module.css';
import line from '~/react/assets/line.svg';
import Amount from '../../helpers/Amount';
import dayjs from 'dayjs';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import {
  getPlanIntervalName,
  getPlanNameWithInterval,
  makePlanSlug
} from '~/react/utils';
import { NEGATIVE_BALANCE_TOOLTIP_MESSAGE } from '~/react/utils/constants';

function LabeledBillingData({
  label,
  value
}: {
  label: string;
  value: string | ReactNode;
}) {
  return (
    <Flex gap="extra-small">
      <Text size={2} weight={500}>
        {label}:
      </Text>
      <Text size={2}>{value}</Text>
    </Flex>
  );
}

function PlanSwitchButton({ nextPlan }: { nextPlan: V1Beta1Plan }) {
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
        variant={'secondary'}
        className={billingStyles.linkBtn}
        onClick={onClick}
        data-test-id="frontier-sdk-plan-switch-btn"
      >
        Switch to {intervalName}
      </Button>
    </div>
  );
}

function getSwitchablePlan(plans: V1Beta1Plan[], currentPlan: V1Beta1Plan) {
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
  // This hook call is duplicated in parent sdks/js/packages/core/react/components/organization/billing/index.tsx
  const {
    client,
    billingAccount,
    config,
    activeSubscription,
    trialSubscription,
    isActiveOrganizationLoading,
    basePlan
  } = useFrontier();
  const [isInvoiceLoading, setIsInvoiceLoading] = useState(false);
  const [memberCount, setMemberCount] = useState(0);
  const [isMemberCountLoading, setIsMemberCountLoading] = useState(false);
  const navigate = useNavigate({ from: '/billing' });

  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plan, setPlan] = useState<V1Beta1Plan>();
  const [switchablePlan, setSwitchablePlan] = useState<V1Beta1Plan | null>(
    null
  );

  useEffect(() => {
    async function getPlans(planId: string) {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceListPlans();
        const plansList = resp?.data?.plans || [];
        const currentPlan = plansList.find(p => p.id === planId);
        setPlan(currentPlan);
        const otherPlan = currentPlan
          ? getSwitchablePlan(plansList, currentPlan)
          : null;
        setSwitchablePlan(otherPlan);
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setIsPlansLoading(false);
      }
    }
    if (activeSubscription?.plan_id) {
      getPlans(activeSubscription?.plan_id);
    }
  }, [client, activeSubscription?.plan_id]);

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
      billingAccount?.org_id &&
      billingAccount?.provider_id
    ) {
      getUpcomingInvoice(billingAccount?.org_id, billingAccount?.id);
      getMemberCount(billingAccount?.org_id);
    }
  }, [
    client,
    billingAccount?.org_id,
    billingAccount?.id,
    billingAccount?.provider_id
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
    phase => phase.plan_id === switchablePlan?.id
  );

  const isLoading =
    isActiveOrganizationLoading ||
    isInvoiceLoading ||
    isMemberCountLoading ||
    isPlansLoading ||
    isPermissionLoading;

  const isUserOnlyTrialing = !activeSubscription?.id && trialSubscription?.id;
  const due_date = upcomingInvoice?.due_date || upcomingInvoice?.period_end_at;

  return isLoading ? (
    <Skeleton />
  ) : due_date && !isUserOnlyTrialing ? (
    <Flex
      align={'center'}
      justify={'between'}
      gap={'small'}
      className={billingStyles.billingCycleBox}
    >
      <Flex gap="medium" align={'center'}>
        <LabeledBillingData label="Plan" value={planName} />
        {switchablePlan && isAllowed && !alreadyPhased ? (
          <PlanSwitchButton
            nextPlan={switchablePlan}
            data-test-id="frontier-sdk-billing-cycle-interval-switch-button"
          />
        ) : null}
      </Flex>
      <Flex gap="medium">
        <LabeledBillingData
          label="Next billing"
          value={dayjs(due_date).format(config.dateFormat)}
        />
        {/* @ts-ignore */}
        <Image src={line} alt="line" />
        <LabeledBillingData label="Users" value={memberCount} />
        {/* @ts-ignore */}
        <Image src={line} alt="line" />
        <LabeledBillingData
          label="Amount"
          value={
            <Flex gap={'medium'}>
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
      align={'center'}
      justify={'between'}
      gap={'small'}
    >
      <Flex gap={'small'}>
        <InfoCircledIcon className={billingStyles.currentPlanInfoText} />
        <Text size={2} className={billingStyles.currentPlanInfoText}>
          {planInfo.message}
        </Text>
      </Flex>
      <Button
        variant={'secondary'}
        onClick={onActionBtnClick}
        data-test-id="frontier-sdk-upcoming-billing-cycle-action-button"
      >
        {planInfo.action.label}
      </Button>
    </Flex>
  );
};
