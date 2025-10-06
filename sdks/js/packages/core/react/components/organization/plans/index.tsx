import { useEffect, useState } from 'react';
import { EmptyState, toast, Skeleton, Text, Flex } from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Feature } from '~/src';
import { groupPlansPricingByInterval } from './helpers';
import { IntervalPricingWithPlan } from '~/src/types';
import { UpcomingPlanChangeBanner } from '~/react/components/common/upcoming-plan-change-banner';
import { PlansHeader } from './header';
import { PlanPricingColumn } from './pricing-column';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import plansStyles from './plans.module.css';
import { styles } from '../styles';
import { Plan } from '@raystack/proton/frontier';

const PlansLoader = () => {
  return (
    <Flex direction="column" gap={4}>
      {[...new Array(2)].map((_, i) => (
        <Skeleton containerClassName={plansStyles.flex1} key={`loader-${i}`} />
      ))}
    </Flex>
  );
};

const NoPlans = () => {
  return (
    <EmptyState
      icon={<ExclamationTriangleIcon />}
      heading={<span style={{ fontWeight: 'bold' }}>No Plans Available</span>}
      subHeading={
        'Sorry, No plans available at this moment. Please try again later'
      }
    />
  );
};

interface PlansListProps {
  plans: Plan[];
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

  const featureTitleMap = features.reduce((acc, f) => {
    const weightage =
      (f.metadata as Record<string, any>)?.weightage || totalFeatures;
    acc[f.title || ''] = weightage;
    return acc;
  }, {} as Record<string, number>);

  const sortedFeatures = Object.entries(featureTitleMap)
    .sort((f1, f2) => f1[1] - f2[1])
    .map(f => f[0])
    .filter(f => Boolean(f));

  return (
    <Flex>
      <Flex style={{ overflow: 'hidden', flex: 1 }}>
        <div className={plansStyles.leftPanel}>
          <div className={plansStyles.planInfoColumn}>{''}</div>
          <Flex direction="column">
            <Flex
              align="center"
              justify="start"
              className={plansStyles.featureCell}
            >
              <Text size="small" className={plansStyles.featureTableHeading}>
                Features
              </Text>
            </Flex>
            {sortedFeatures.map(feature => {
              return (
                <Flex
                  key={feature}
                  align="center"
                  justify="start"
                  className={plansStyles.featureCell}
                >
                  <Text size={3} className={plansStyles.featureLabel}>
                    {feature}
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
  const {
    config,
    client,
    activeSubscription,
    isActiveSubscriptionLoading,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();
  const [features, setFeatures] = useState<V1Beta1Feature[]>([]);

  const { isFetching: isPermissionsFetching, isAllowed: canChangePlan } =
    useBillingPermission();

  useEffect(() => {
    async function getFeatures() {
      try {
        const featuresResp = await client?.frontierServiceListFeatures();
        if (featuresResp?.data?.features) {
          setFeatures(featuresResp?.data?.features);
        }
      } catch (err: any) {
        toast.error('Something went wrong', {
          description: err.message
        });
        console.error(err);
      }
    }

    getFeatures();
  }, [client]);

  const plans = [...(basePlan ? [basePlan] : []), ...allPlans];

  const isLoading =
    isAllPlansLoading ||
    isPermissionsFetching ||
    isActiveSubscriptionLoading ||
    isActiveOrganizationLoading;

  return (
    <Flex direction="column" style={{ width: '100%', overflow: 'hidden' }}>
      <Flex style={styles.header}>
        <Text size="large">Plans</Text>
      </Flex>
      <Flex direction="column" gap={7} style={{ ...styles.container }}>
        <Flex direction="column">
          <PlansHeader billingSupportEmail={config.billing?.supportEmail} />
        </Flex>
        <UpcomingPlanChangeBanner
          isLoading={isLoading}
          subscription={activeSubscription}
          isAllowed={canChangePlan}
        />
        {isLoading ? (
          <PlansLoader />
        ) : (
          <PlansList
            plans={plans}
            features={features}
            currentPlanId={activeSubscription?.planId || ''}
            allowAction={canChangePlan}
          />
        )}
      </Flex>
      <Outlet />
    </Flex>
  );
}
