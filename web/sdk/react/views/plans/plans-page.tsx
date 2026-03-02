import { useState } from 'react';
import { EmptyState, Skeleton, Text, Flex } from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { groupPlansPricingByInterval } from './helpers';
import { IntervalPricingWithPlan } from '~/src/types';
import { UpcomingPlanChangeBanner } from '~/react/components/common/upcoming-plan-change-banner';
import { PlansHeader } from './plans-header';
import { PlanPricingColumn } from './pricing-column';
import { ConfirmPlanChangeDialog } from './confirm-plan-change-dialog';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { useQuery as useConnectQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '~hooks';
import { create } from '@bufbuild/protobuf';
import { Feature, ListFeaturesRequestSchema } from '@raystack/proton/frontier';
import { Plan } from '@raystack/proton/frontier';
import sharedStyles from '../../components/organization/styles.module.css';
import plansStyles from './plans.module.css';

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
  features: Feature[];
  onConfirmPlanChange?: (planId: string) => void;
}

const PlansList = ({
  plans = [],
  features = [],
  currentPlanId,
  allowAction,
  onConfirmPlanChange
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
              onConfirmPlanChange={onConfirmPlanChange}
            />
          ))}
        </Flex>
      </Flex>
    </Flex>
  );
};

export interface PlansPageProps {}

export default function PlansPage(_props: PlansPageProps = {}) {
  const {
    config,
    activeSubscription,
    isActiveSubscriptionLoading,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading
  } = useFrontier();

  const { isFetching: isPermissionsFetching, isAllowed: canChangePlan } =
    useBillingPermission();

  const { data: featuresData } = useConnectQuery(
    FrontierServiceQueries.listFeatures,
    create(ListFeaturesRequestSchema, {})
  );

  const features = (featuresData?.features || []) as Feature[];

  const plans = [...(basePlan ? [basePlan] : []), ...allPlans];

  const isLoading =
    isAllPlansLoading ||
    isPermissionsFetching ||
    isActiveSubscriptionLoading ||
    isActiveOrganizationLoading;

  const [confirmPlanChangeState, setConfirmPlanChangeState] = useState({
    open: false,
    planId: ''
  });

  const handleConfirmPlanChangeOpenChange = (value: boolean) => {
    if (!value) {
      setConfirmPlanChangeState({ open: false, planId: '' });
    } else {
      setConfirmPlanChangeState(prev => ({ ...prev, open: value }));
    }
  };

  return (
    <Flex direction="column" style={{ width: '100%', overflow: 'hidden' }}>
      <Flex direction="column" className={sharedStyles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <PlansHeader
            billingSupportEmail={config.billing?.supportEmail}
            isLoading={isLoading}
          />
        </Flex>
        <Flex direction="column" gap={7}>
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
            onConfirmPlanChange={(planId) =>
              setConfirmPlanChangeState({ open: true, planId })
            }
          />
        )}
        </Flex>
      </Flex>
      <ConfirmPlanChangeDialog
        open={confirmPlanChangeState.open}
        onOpenChange={handleConfirmPlanChangeOpenChange}
        planId={confirmPlanChangeState.planId}
      />
    </Flex>
  );
}
