import { useNavigate } from '@tanstack/react-router';
import { ReactNode, useEffect, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Invoice, V1Beta1Plan } from '~/src';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import { Flex, Text, Image, Button } from '@raystack/apsara';
import billingStyles from './billing.module.css';
import line from '~/react/assets/line.svg';
import Amount from '../../helpers/Amount';
import dayjs from 'dayjs';
import { InfoCircledIcon } from '@radix-ui/react-icons';

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

export const UpcomingBillingCycle = () => {
  const [upcomingInvoice, setUpcomingInvoice] = useState<V1Beta1Invoice>();
  const {
    client,
    billingAccount,
    config,
    activeSubscription,
    isActiveOrganizationLoading
  } = useFrontier();
  const [isInvoiceLoading, setIsInvoiceLoading] = useState(false);
  const [memberCount, setMemberCount] = useState(0);
  const [isMemberCountLoading, setIsMemberCountLoading] = useState(false);
  const navigate = useNavigate({ from: '/billing' });

  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plan, setPlan] = useState<V1Beta1Plan>();

  useEffect(() => {
    async function getPlan(planId: string) {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceGetPlan(planId);
        if (resp?.data?.plan) {
          setPlan(resp?.data?.plan);
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
    if (activeSubscription?.plan_id) {
      getPlan(activeSubscription?.plan_id);
    }
  }, [client, activeSubscription?.plan_id]);

  useEffect(() => {
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

    if (billingAccount?.id && billingAccount?.org_id) {
      getUpcomingInvoice(billingAccount?.org_id, billingAccount?.id);
    }
  }, [client, billingAccount?.org_id, billingAccount?.id]);

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
        if (invoice) {
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

    if (billingAccount?.id && billingAccount?.org_id) {
      getUpcomingInvoice(billingAccount?.org_id, billingAccount?.id);
      getMemberCount(billingAccount?.org_id);
    }
  }, [client, billingAccount?.org_id, billingAccount?.id]);

  const planName = plan?.title;

  const planInfo = activeSubscription
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

  const isLoading =
    isActiveOrganizationLoading ||
    isInvoiceLoading ||
    isMemberCountLoading ||
    isPlansLoading;

  return isLoading ? (
    <Skeleton />
  ) : upcomingInvoice ? (
    <Flex
      align={'center'}
      justify={'between'}
      gap={'small'}
      className={billingStyles.billingCycleBox}
    >
      <LabeledBillingData label="Plan" value={planName} />
      <Flex gap="medium">
        <LabeledBillingData
          label="Next billing"
          value={dayjs(upcomingInvoice?.due_date).format(config.dateFormat)}
        />
        {/* @ts-ignore */}
        <Image src={line} alt="line" />
        <LabeledBillingData label="Users" value={memberCount} />
        {/* @ts-ignore */}
        <Image src={line} alt="line" />
        <LabeledBillingData
          label="Amount"
          value={
            <Amount
              currency={upcomingInvoice?.currency}
              value={Number(upcomingInvoice?.amount)}
            />
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
      <Button variant={'secondary'} onClick={onActionBtnClick}>
        {planInfo.action.label}
      </Button>
    </Flex>
  );
};
