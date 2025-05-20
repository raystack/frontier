import { EmptyState, Flex } from '@raystack/apsara';
import { Text } from '@raystack/apsara/v1';
import { styles } from '../styles';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useEffect, useState } from 'react';
import { V1Beta1Feature, V1Beta1Plan } from '~/src';
import { toast } from '@raystack/apsara/v1';
import Skeleton from 'react-loading-skeleton';
import plansStyles from './plans.module.css';
import { groupPlansPricingByInterval } from './helpers';
import { IntervalPricingWithPlan } from '~/src/types';

import { Outlet } from '@tanstack/react-router';

import { UpcomingPlanChangeBanner } from '~/react/components/common/upcoming-plan-change-banner';
import { PlansHeader } from './header';
import { PlanPricingColumn } from './pricing-column';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';

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
      <Text size="large" weight="medium">
        No Plans Available
      </Text>
      <Text size="small">
        Sorry, No plans available at this moment. Please try again later
      </Text>
    </EmptyState>
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
          <Flex direction={'column'}>
            <Flex
              align={'center'}
              justify={'start'}
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
                  align={'center'}
                  justify={'start'}
                  className={plansStyles.featureCell}
                >
                  <Text size="regular" className={plansStyles.featureLabel}>
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
    basePlan
  } = useFrontier();
  const [isPlansLoading, setIsPlansLoading] = useState(false);
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [features, setFeatures] = useState<V1Beta1Feature[]>([]);

  const { isFetching: isPermissionsFetching, isAllowed: canChangePlan } =
    useBillingPermission();

  useEffect(() => {
    async function getPlansAndFeatures() {
      setIsPlansLoading(true);
      try {
        const [planResp, featuresResp] = await Promise.all([
          client?.frontierServiceListPlans(),
          client?.frontierServiceListFeatures()
        ]);
        if (planResp?.data?.plans) {
          setPlans([...(basePlan ? [basePlan] : []), ...planResp?.data?.plans]);
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
  }, [client, basePlan]);

  const isLoading =
    isPlansLoading ||
    isPermissionsFetching ||
    isActiveSubscriptionLoading ||
    isActiveOrganizationLoading;

  return (
    <Flex direction="column" style={{ width: '100%', overflow: 'hidden' }}>
      <Flex style={styles.header}>
        <Text size="large" weight="medium">Plans</Text>
      </Flex>
      <Flex direction="column" style={{ ...styles.container, gap: '24px' }}>
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
            currentPlanId={activeSubscription?.plan_id || ''}
            allowAction={canChangePlan}
          />
        )}
      </Flex>
      <Outlet />
    </Flex>
  );
}
