import { Button, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useState } from 'react';
import billingStyles from './billing.module.css';
import {
  V1Beta1BillingAccount,
  V1Beta1PaymentMethod,
  V1Beta1Plan,
  V1Beta1Subscription
} from '~/src';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import * as _ from 'lodash';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import { converBillingAddressToString } from '~/react/utils';
import Invoices from './invoices';

interface BillingHeaderProps {
  billingSupportEmail?: string;
}

const BillingHeader = ({ billingSupportEmail }: BillingHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>Billing</Text>
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

interface BillingDetailsProps {
  billingAccount?: V1Beta1BillingAccount;
  onAddDetailsClick?: () => void;
  isLoading: boolean;
}

const BillingDetails = ({
  billingAccount,
  onAddDetailsClick = () => {},
  isLoading
}: BillingDetailsProps) => {
  const addressStr = converBillingAddressToString(billingAccount?.address);
  const btnText = addressStr && billingAccount?.name ? 'Update' : 'Add details';
  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Billing Details</Text>
        <Button variant={'secondary'} onClick={onAddDetailsClick}>
          {btnText}
        </Button>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Name</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : billingAccount?.name || 'N/A'}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Address</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton count={2} /> : addressStr || 'N/A'}
        </Text>
      </Flex>
    </div>
  );
};

interface PaymentMethodProps {
  paymentMethod?: V1Beta1PaymentMethod;
  isLoading: boolean;
}

const PaymentMethod = ({
  paymentMethod = {},
  isLoading
}: PaymentMethodProps) => {
  const {
    card_last4 = '',
    card_expiry_month,
    card_expiry_year
  } = paymentMethod;
  // TODO: change card digit as per card type
  const cardDigit = 12;
  const cardNumber = card_last4 ? _.repeat('*', cardDigit) + card_last4 : 'N/A';
  const cardExp =
    card_expiry_month && card_expiry_year
      ? `${card_expiry_month}/${card_expiry_year}`
      : 'N/A';

  return (
    <div className={billingStyles.detailsBox}>
      <Flex align={'center'} justify={'between'} style={{ width: '100%' }}>
        <Text className={billingStyles.detailsBoxHeading}>Payment method</Text>
        <Button variant={'secondary'}>Add method</Button>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Card Number</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : cardNumber}
        </Text>
      </Flex>
      <Flex direction={'column'} gap={'extra-small'}>
        <Text className={billingStyles.detailsBoxRowLabel}>Expiry</Text>
        <Text className={billingStyles.detailsBoxRowValue}>
          {isLoading ? <Skeleton /> : cardExp}
        </Text>
      </Flex>
    </div>
  );
};

interface CurrentPlanInfoProps {
  subscription?: V1Beta1Subscription;
  isLoading?: boolean;
}

const CurrentPlanInfo = ({ subscription, isLoading }: CurrentPlanInfoProps) => {
  const navigate = useNavigate({ from: '/billing' });
  const { client } = useFrontier();
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
    if (subscription?.plan_id) {
      getPlan(subscription?.plan_id);
    }
  }, [client, subscription?.plan_id]);

  const planName = plan?.title;

  const planInfo = subscription
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

  const showLoader = isLoading || isPlansLoading;

  return showLoader ? (
    <Skeleton />
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

export default function Billing() {
  const {
    billingAccount: activeBillingAccount,
    activeSubscription,
    isActiveSubscriptionLoading,
    client,
    config
  } = useFrontier();
  const navigate = useNavigate({ from: '/billing' });
  const [billingAccount, setBillingAccount] = useState<V1Beta1BillingAccount>();
  const [paymentMethod, setPaymentMethod] = useState<V1Beta1PaymentMethod>();
  const [isBillingAccountLoading, setBillingAccountLoading] = useState(false);

  useEffect(() => {
    async function getPaymentMethod(orgId: string, billingId: string) {
      setBillingAccountLoading(true);
      try {
        const resp = await client?.frontierServiceGetBillingAccount(
          orgId,
          billingId,
          { with_payment_methods: true }
        );
        if (resp?.data) {
          const paymentMethods = resp?.data?.payment_methods || [];
          setBillingAccount(resp.data.billing_account);
          setPaymentMethod(paymentMethods[0]);
        }
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      } finally {
        setBillingAccountLoading(false);
      }
    }

    if (activeBillingAccount?.id && activeBillingAccount?.org_id) {
      getPaymentMethod(activeBillingAccount?.org_id, activeBillingAccount?.id);
    }
  }, [activeBillingAccount?.id, activeBillingAccount?.org_id, client]);

  const onAddDetailsClick = useCallback(() => {
    if (billingAccount?.id) {
      navigate({
        to: '/billing/$billingId/edit-address',
        params: { billingId: billingAccount?.id }
      });
    }
  }, [billingAccount?.id, navigate]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Billing</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <BillingHeader billingSupportEmail={config.billing?.supportEmail} />
          <Flex style={{ gap: '24px' }}>
            <PaymentMethod
              paymentMethod={paymentMethod}
              isLoading={isBillingAccountLoading}
            />
            <BillingDetails
              billingAccount={billingAccount}
              onAddDetailsClick={onAddDetailsClick}
              isLoading={isBillingAccountLoading}
            />
          </Flex>
          <CurrentPlanInfo
            subscription={activeSubscription}
            isLoading={isActiveSubscriptionLoading}
          />
          <Invoices
            organizationId={activeBillingAccount?.org_id || ''}
            billingId={activeBillingAccount?.id || ''}
            isLoading={isBillingAccountLoading}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
