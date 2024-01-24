import {
  Button,
  EmptyState,
  Flex,
  Text,
  ToggleGroup,
  Image
} from '@raystack/apsara';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { V1Beta1Feature, V1Beta1Plan } from '~/src';
import { toast } from 'sonner';
import Skeleton from 'react-loading-skeleton';
import plansStyles from './plans.module.css';
import { getAllPlansFeatuesMap, groupPlansPricingByInterval } from './helpers';
import {
  IntervalKeys,
  IntervalLabelMap,
  PlanIntervalPricing
} from '~/src/types';
import checkCircle from '~/react/assets/check-circle.svg';
import qs from 'query-string';

const PlansLoader = () => {
  return (
    <Flex direction={'column'}>
      {[...new Array(15)].map((_, i) => (
        <Skeleton containerClassName={plansStyles.flex1} key={`loader-${i}`} />
      ))}
    </Flex>
  );
};

const NoPlans = () => {
  return (
    <EmptyState style={{ marginTop: 160 }}>
      <Text size={5} style={{ fontWeight: 'bold' }}>
        No Plans Available
      </Text>
      <Text size={2}>
        Sorry, No plans available at this moment. Please try again later
      </Text>
    </EmptyState>
  );
};

interface PlansHeaderProps {
  billingSupportEmail?: string;
}

const PlansHeader = ({ billingSupportEmail }: PlansHeaderProps) => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>Plans</Text>
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

const PlanPricingColumn = ({
  plan,
  featureMap = {}
}: {
  plan: PlanIntervalPricing;
  featureMap: Record<string, V1Beta1Feature>;
}) => {
  const {
    client,
    activeOrganization,
    billingAccount,
    config,
    activeSubscription
  } = useFrontier();

  const [isLoading, setIsLoading] = useState(false);

  const planIntervals = (Object.keys(plan.intervals).sort() ||
    []) as IntervalKeys[];
  const [selectedInterval, setSelectedInterval] = useState<IntervalKeys>(() => {
    const activePlan = Object.values(plan?.intervals).find(
      p => p.planId === activeSubscription?.plan_id
    );
    return activePlan?.interval || planIntervals[0];
  });

  const onIntervalChange = (value: IntervalKeys) => {
    if (value) {
      setSelectedInterval(value);
    }
  };

  const selectedIntervalPricing = plan.intervals[selectedInterval];

  const action = useMemo(() => {
    if (selectedIntervalPricing.planId === activeSubscription?.plan_id) {
      return {
        disabled: true,
        text: 'Current Plan'
      };
    }
    return {
      disabled: false,
      text: 'Upgrade'
    };
  }, [activeSubscription?.plan_id, selectedIntervalPricing.planId]);

  const onPlanActionClick = useCallback(async () => {
    setIsLoading(true);
    try {
      if (activeOrganization?.id && billingAccount?.id) {
        const query = qs.stringify(
          {
            details: btoa(
              qs.stringify({
                billing_id: billingAccount?.id,
                organization_id: activeOrganization?.id,
                type: 'plans'
              })
            ),
            checkout_id: '{{.CheckoutID}}'
          },
          { encode: false }
        );
        const cancel_url = `${config?.billing?.cancelUrl}?${query}`;
        const success_url = `${config?.billing?.successUrl}?${query}`;

        const resp = await client?.frontierServiceCreateCheckout(
          activeOrganization?.id,
          billingAccount?.id,
          {
            cancel_url: cancel_url,
            success_url: success_url,
            subscription_body: {
              plan: selectedIntervalPricing?.planId
            }
          }
        );
        if (resp?.data?.checkout_session?.checkout_url) {
          window.location.href = resp?.data?.checkout_session?.checkout_url;
        }
      }
    } catch (err: any) {
      console.error(err);
      toast.error('Something went wrong', {
        description: err?.message
      });
    } finally {
      setIsLoading(false);
    }
  }, [
    activeOrganization?.id,
    billingAccount?.id,
    config?.billing?.cancelUrl,
    config?.billing?.successUrl,
    client,
    selectedIntervalPricing?.planId
  ]);

  return (
    <Flex direction={'column'} style={{ flex: 1 }}>
      <Flex className={plansStyles.planInfoColumn} direction="column">
        <Flex gap="small" direction="column">
          <Text size={4} className={plansStyles.planTitle}>
            {plan.title}
          </Text>
          <Flex gap={'extra-small'} align={'end'}>
            <Text size={8} className={plansStyles.planPrice}>
              {selectedIntervalPricing.currency}{' '}
              {selectedIntervalPricing.amount?.toString()}
            </Text>
            <Text size={2} className={plansStyles.planPriceSub}>
              per seat/{selectedInterval}
            </Text>
          </Flex>
          <Text size={2} className={plansStyles.planDescription}>
            {plan?.description}
          </Text>
        </Flex>
        <Flex direction="column" gap="medium">
          <Button
            variant={'secondary'}
            className={plansStyles.planActionBtn}
            onClick={onPlanActionClick}
            disabled={action?.disabled || isLoading}
          >
            {isLoading ? 'Upgrading...' : action.text}
          </Button>
          {planIntervals.length > 1 ? (
            <ToggleGroup
              className={plansStyles.plansIntervalList}
              value={selectedInterval}
              onValueChange={onIntervalChange}
            >
              {planIntervals.map(key => (
                <ToggleGroup.Item
                  value={key}
                  key={key}
                  className={plansStyles.plansIntervalListItem}
                >
                  <Text className={plansStyles.plansIntervalListItemText}>
                    {IntervalLabelMap[key]}
                  </Text>
                </ToggleGroup.Item>
              ))}
            </ToggleGroup>
          ) : null}
        </Flex>
      </Flex>
      <Flex direction={'column'}>
        <Flex
          align={'center'}
          justify={'start'}
          className={plansStyles.featureCell}
        >
          <Text size={2} className={plansStyles.featureTableHeading}>
            {plan.title}
          </Text>
        </Flex>
      </Flex>
      {Object.values(featureMap).map(feature => {
        return (
          <Flex
            key={feature?.id + '-' + plan?.slug}
            align={'center'}
            justify={'start'}
            className={plansStyles.featureCell}
          >
            {plan?.features.hasOwnProperty(feature?.id || '') ? (
              <Image
                // @ts-ignore
                src={checkCircle}
                alt="checked"
              />
            ) : (
              ''
            )}
          </Flex>
        );
      })}
    </Flex>
  );
};

interface PlansListProps {
  plans: V1Beta1Plan[];
}

const PlansList = ({ plans = [] }: PlansListProps) => {
  if (plans.length === 0) return <NoPlans />;

  const groupedPlans = groupPlansPricingByInterval(plans);
  const featuresMap = getAllPlansFeatuesMap(plans);
  return (
    <Flex>
      <Flex style={{ overflow: 'hidden', flex: 1 }}>
        <div className={plansStyles.leftPanel}>
          <div className={plansStyles.planInfoColumn}>{''}</div>
          <Flex direction={'column'}>
            <Flex
              align={'center'}
              justify={'start'}
              className={plansStyles.featureCell}
            >
              <Text size={2} className={plansStyles.featureTableHeading}>
                Features
              </Text>
            </Flex>
            {Object.values(featuresMap).map(feature => {
              return (
                <Flex
                  key={feature?.id}
                  align={'center'}
                  justify={'start'}
                  className={plansStyles.featureCell}
                >
                  <Text size={3} className={plansStyles.featureLabel}>
                    {feature?.name}
                  </Text>
                </Flex>
              );
            })}
          </Flex>
        </div>
        <Flex className={plansStyles.rightPanel}>
          {groupedPlans.map(plan => (
            <PlanPricingColumn
              plan={plan}
              key={plan.slug}
              featureMap={featuresMap}
            />
          ))}
        </Flex>
      </Flex>
    </Flex>
  );
};

export default function Plans() {
  const { config, client } = useFrontier();
  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);

  useEffect(() => {
    async function getPlans() {
      setIsPlansLoading(true);
      try {
        const resp = await client?.frontierServiceListPlans();
        if (resp?.data?.plans) {
          setPlans(resp?.data?.plans);
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

    getPlans();
  }, [client]);

  return (
    <Flex direction="column" style={{ width: '100%', overflow: 'hidden' }}>
      <Flex style={styles.header}>
        <Text size={6}>Plans</Text>
      </Flex>
      <Flex direction="column" style={{ ...styles.container, gap: '24px' }}>
        <Flex direction="column">
          <PlansHeader billingSupportEmail={config.billing?.supportEmail} />
        </Flex>
        {isPlansLoading ? <PlansLoader /> : <PlansList plans={plans} />}
      </Flex>
    </Flex>
  );
}
