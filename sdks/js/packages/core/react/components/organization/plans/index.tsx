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
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import checkCircle from '~/react/assets/check-circle.svg';
import { PlanChangeAction, getPlanChangeAction } from '~/react/utils';
import Amount from '../../helpers/Amount';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { usePlans } from './hooks/usePlans';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';

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
  featureMap = {},
  currentPlan,
  allowAction
}: {
  plan: PlanIntervalPricing;
  featureMap: Record<string, V1Beta1Feature>;
  currentPlan?: IntervalPricingWithPlan;
  allowAction: boolean;
}) => {
  const { config } = useFrontier();

  const navigate = useNavigate({ from: '/plans' });

  const { checkoutPlan, isLoading } = usePlans();

  const planIntervals =
    Object.values(plan.intervals)
      .sort((a, b) => a.weightage - b.weightage)
      .map(i => i.interval) || [];

  const [selectedInterval, setSelectedInterval] = useState<IntervalKeys>(() => {
    const activePlan = Object.values(plan?.intervals).find(
      p => p.planId === currentPlan?.planId
    );
    return activePlan?.interval || planIntervals[0];
  });

  const onIntervalChange = (value: IntervalKeys) => {
    if (value) {
      setSelectedInterval(value);
    }
  };

  const selectedIntervalPricing = plan.intervals[selectedInterval];

  const action: PlanChangeAction = useMemo(() => {
    if (selectedIntervalPricing.planId === currentPlan?.planId) {
      return {
        disabled: true,
        btnLabel: 'Current Plan',
        btnLoadingLabel: 'Current Plan',
        btnDoneLabel: ''
      };
    }

    const planAction = getPlanChangeAction(
      selectedIntervalPricing.weightage,
      currentPlan?.weightage
    );
    return {
      disabled: false,
      ...planAction
    };
  }, [currentPlan, selectedIntervalPricing]);

  const onPlanActionClick = useCallback(() => {
    if (action?.showModal) {
      navigate({
        to: '/plans/confirm-change/$planId',
        params: {
          planId: selectedIntervalPricing?.planId
        }
      });
    } else {
      checkoutPlan({
        planId: selectedIntervalPricing?.planId,
        onSuccess: data => {
          window.location.href = data?.checkout_url as string;
        }
      });
    }
  }, [
    action?.showModal,
    checkoutPlan,
    navigate,
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
            <Amount
              value={selectedIntervalPricing.amount}
              currency={selectedIntervalPricing.currency}
              className={plansStyles.planPrice}
              hideDecimals={config?.billing?.hideDecimals}
            />
            <Text size={2} className={plansStyles.planPriceSub}>
              per seat/{selectedInterval}
            </Text>
          </Flex>
          <Text size={2} className={plansStyles.planDescription}>
            {plan?.description}
          </Text>
        </Flex>
        <Flex direction="column" gap="medium">
          {allowAction ? (
            <Button
              variant={'secondary'}
              className={plansStyles.planActionBtn}
              onClick={onPlanActionClick}
              disabled={action?.disabled || isLoading}
            >
              {isLoading ? `${action.btnLoadingLabel}....` : action.btnLabel}
            </Button>
          ) : null}
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
  currentPlanId: string;
  allowAction: boolean;
}

const PlansList = ({
  plans = [],
  currentPlanId,
  allowAction
}: PlansListProps) => {
  if (plans.length === 0) return <NoPlans />;

  const groupedPlans = groupPlansPricingByInterval(plans).sort(
    (a, b) => a.weightage - b.weightage
  );
  const featuresMap = getAllPlansFeatuesMap(plans);

  let currentPlanPricing: IntervalPricingWithPlan | undefined;
  groupedPlans.forEach(group => {
    Object.values(group.intervals).forEach(plan => {
      if (plan.planId === currentPlanId) {
        currentPlanPricing = plan;
      }
    });
  });

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
              currentPlan={currentPlanPricing}
              allowAction={allowAction}
            />
          ))}
        </Flex>
      </Flex>
    </Flex>
  );
};

export default function Plans() {
  const { config, client, activeSubscription, activeOrganization } =
    useFrontier();
  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);

  const resource = `app/organization:${activeOrganization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!activeOrganization?.id
  );

  const { canChangePlan } = useMemo(() => {
    return {
      canChangePlan: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

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

  const isLoading = isPlansLoading || isPermissionsFetching;

  return (
    <Flex direction="column" style={{ width: '100%', overflow: 'hidden' }}>
      <Flex style={styles.header}>
        <Text size={6}>Plans</Text>
      </Flex>
      <Flex direction="column" style={{ ...styles.container, gap: '24px' }}>
        <Flex direction="column">
          <PlansHeader billingSupportEmail={config.billing?.supportEmail} />
        </Flex>
        {isLoading ? (
          <PlansLoader />
        ) : (
          <PlansList
            plans={plans}
            currentPlanId={activeSubscription?.plan_id || ''}
            allowAction={canChangePlan}
          />
        )}
      </Flex>
      <Outlet />
    </Flex>
  );
}
