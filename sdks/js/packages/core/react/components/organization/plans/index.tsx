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
import { groupPlansPricingByInterval } from './helpers';
import {
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import checkCircle from '~/react/assets/check-circle.svg';
import { PlanChangeAction, getPlanChangeAction } from '~/react/utils';
import Amount from '~/react/components/helpers/Amount';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { usePlans } from './hooks/usePlans';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';
import * as _ from 'lodash';
import dayjs from 'dayjs';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { UpcomingPlanChangeBanner } from '~/react/components/common/upcoming-plan-change-banner';

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
  features,
  currentPlan,
  allowAction
}: {
  plan: PlanIntervalPricing;
  features: V1Beta1Feature[];
  currentPlan?: IntervalPricingWithPlan;
  allowAction: boolean;
}) => {
  const { config } = useFrontier();

  const navigate = useNavigate({ from: '/plans' });

  const { checkoutPlan, isLoading, changePlan, verifyPlanChange } = usePlans();

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

  const isAlreadySubscribed = !_.isEmpty(currentPlan);

  const onPlanActionClick = useCallback(() => {
    if (action?.showModal) {
      navigate({
        to: '/plans/confirm-change/$planId',
        params: {
          planId: selectedIntervalPricing?.planId
        }
      });
    } else if (isAlreadySubscribed) {
      const planId = selectedIntervalPricing?.planId;
      changePlan({
        planId,
        onSuccess: async () => {
          const planPhase = await verifyPlanChange({ planId });
          if (planPhase) {
            const changeDate = dayjs(planPhase?.effective_at).format(
              config?.dateFormat || DEFAULT_DATE_FORMAT
            );
            const actionName = action?.btnLabel.toLowerCase();
            toast.success(`Plan ${actionName} successful`, {
              description: `Your plan will ${actionName} on ${changeDate}`
            });
          }
        },
        immediate: action?.immediate
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
    action?.immediate,
    action?.btnLabel,
    isAlreadySubscribed,
    navigate,
    selectedIntervalPricing?.planId,
    changePlan,
    verifyPlanChange,
    config?.dateFormat,
    checkoutPlan
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
      {features.map(feature => {
        const featureId = feature?.id || '';
        const planFeature = _.get(plan?.features, featureId, { metadata: {} });
        const value = (planFeature?.metadata as Record<string, any>)?.value;
        const isAvailable = value?.toLowerCase() === 'true';
        return (
          <Flex
            key={featureId + '-' + plan?.slug}
            align={'center'}
            justify={'start'}
            className={plansStyles.featureCell}
          >
            {isAvailable ? (
              <Image
                // @ts-ignore
                src={checkCircle}
                alt="checked"
              />
            ) : value ? (
              <Text>{value}</Text>
            ) : null}
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
  features: V1Beta1Feature[];
}

const PlansList = ({
  plans = [],
  features = [],
  currentPlanId,
  allowAction
}: PlansListProps) => {
  if (plans.length === 0) return <NoPlans />;

  const groupedPlans = groupPlansPricingByInterval(plans).sort(
    (a, b) => a.weightage - b.weightage
  );

  let currentPlanPricing: IntervalPricingWithPlan | undefined;
  groupedPlans.forEach(group => {
    Object.values(group.intervals).forEach(plan => {
      if (plan.planId === currentPlanId) {
        currentPlanPricing = plan;
      }
    });
  });

  const totalFeatures = features.length;
  const sortedFeatures = features
    .sort((f1, f2) => ((f1?.name || '') > (f2?.name || '') ? -1 : 1))
    .sort((f1, f2) => {
      const f1Weight =
        (f1.metadata as Record<string, any>)?.weightage || totalFeatures;
      const f2Weight =
        (f2.metadata as Record<string, any>)?.weightage || totalFeatures;
      return f1Weight - f2Weight;
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
            {sortedFeatures.map(feature => {
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
              features={sortedFeatures}
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
  const [features, setFeatures] = useState<V1Beta1Feature[]>([]);

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
    async function getPlansAndFeatures() {
      setIsPlansLoading(true);
      try {
        const [planResp, featuresResp] = await Promise.all([
          client?.frontierServiceListPlans(),
          client?.frontierServiceListFeatures()
        ]);
        if (planResp?.data?.plans) {
          setPlans(planResp?.data?.plans);
        }
        if (featuresResp?.data?.features) {
          setFeatures(featuresResp?.data?.features);
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

    getPlansAndFeatures();
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
        <UpcomingPlanChangeBanner
          isLoading={isLoading}
          subscription={activeSubscription}
        />
        {isLoading ? (
          <PlansLoader />
        ) : (
          <PlansList
            plans={plans}
            features={features}
            currentPlanId={activeSubscription?.plan_id || ''}
            allowAction={canChangePlan}
          />
        )}
      </Flex>
      <Outlet />
    </Flex>
  );
}
