'use client';

import { useMemo, useState } from 'react';
import {
  Dialog,
  EmptyState,
  Flex,
  Skeleton
} from '@raystack/apsara-v1';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  Feature,
  FrontierServiceQueries,
  ListFeaturesRequestSchema
} from '@raystack/proton/frontier';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useBillingPermission } from '~/react/hooks/useBillingPermission';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import { UpcomingPlanChangeBanner } from '~/react/views-new/billing/components/upcoming-plan-change-banner';
import { groupPlansPricingByInterval } from './helpers';
import { IntervalKeys, IntervalPricingWithPlan } from '~/src/types';
import { PlanCard } from './components/plan-card';
import { FeatureTable } from './components/feature-table';
import {
  ConfirmPlanChangeDialog,
  type ConfirmPlanChangePayload
} from './components/confirm-plan-change-dialog';
import styles from './plans-view.module.css';

const confirmPlanChangeHandle =
  Dialog.createHandle<ConfirmPlanChangePayload>();

export function PlansView() {
  const {
    config,
    activeSubscription,
    isActiveSubscriptionLoading,
    isActiveOrganizationLoading,
    basePlan,
    allPlans,
    isAllPlansLoading,
    activeOrganization,
  } = useFrontier();

  const { isFetching: isPermissionsFetching, isAllowed: canChangePlan } =
    useBillingPermission();

  const { data: featuresData, isLoading: isFeaturesLoading } = useQuery(
    FrontierServiceQueries.listFeatures,
    create(ListFeaturesRequestSchema, {})
  );

  const features = useMemo(
    () => (featuresData?.features || []) as Feature[],
    [featuresData]
  );

  const plans = useMemo(
    () => [...(basePlan ? [basePlan] : []), ...allPlans],
    [basePlan, allPlans]
  );

  const groupedPlans = useMemo(
    () =>
      groupPlansPricingByInterval(plans).sort(
        (a, b) => a.weightage - b.weightage
      ),
    [plans]
  );

  const isLoading = !activeOrganization?.id ||
    isAllPlansLoading ||
    isPermissionsFetching ||
    isActiveSubscriptionLoading ||
    isActiveOrganizationLoading;

  const totalFeatures = features.length;

  const sortedFeatures = useMemo(() => {
    const featureTitleMap = features.reduce(
      (acc, f) => {
        const weightage =
          (f.metadata as Record<string, number>)?.weightage || totalFeatures;
        acc[f.title || ''] = weightage;
        return acc;
      },
      {} as Record<string, number>
    );

    return Object.entries(featureTitleMap)
      .sort((f1, f2) => f1[1] - f2[1])
      .map(f => f[0])
      .filter(Boolean);
  }, [features, totalFeatures]);

  const [intervalOverrides, setIntervalOverrides] = useState<
    Record<string, IntervalKeys>
  >({});

  const selectedIntervals = useMemo<Record<string, IntervalKeys>>(() => {
    const result: Record<string, IntervalKeys> = {};
    groupedPlans.forEach(plan => {
      if (intervalOverrides[plan.slug]) {
        result[plan.slug] = intervalOverrides[plan.slug];
        return;
      }
      const sortedIntervals = Object.values(plan.intervals)
        .sort((a, b) => a.weightage - b.weightage)
        .map(i => i.interval);
      const activePlanInterval = Object.values(plan.intervals).find(
        p => p.planId === activeSubscription?.planId
      );
      result[plan.slug] =
        activePlanInterval?.interval || sortedIntervals[0] || 'year';
    });
    return result;
  }, [groupedPlans, activeSubscription?.planId, intervalOverrides]);

  let currentPlanPricing: IntervalPricingWithPlan | undefined;
  groupedPlans.forEach(group => {
    Object.values(group.intervals).forEach(plan => {
      if (plan.planId === activeSubscription?.planId) {
        currentPlanPricing = plan;
      }
    });
  });

  const billingSupportEmail = config.billing?.supportEmail;

  const description = billingSupportEmail
    ? `View and manage your subscription plan. For more details, contact ${billingSupportEmail}`
    : 'View and manage your subscription plan.';

  if (isLoading) {
    return (
      <ViewContainer>
        <ViewHeader title="Plans" description="View and manage your subscription plan." />
        <Flex direction="column" gap={4}>
          <Skeleton height="200px" />
          <Skeleton height="400px" />
        </Flex>
      </ViewContainer>
    );
  }

  if (plans.length === 0) {
    return (
      <ViewContainer>
        <ViewHeader title="Plans" description="View and manage your subscription plan." />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="No Plans Available"
          subHeading="No plans available at this moment. Please try again later."
        />
      </ViewContainer>
    );
  }

  return (
    <ViewContainer>
      <ViewHeader title="Plans" description={description} />
      <UpcomingPlanChangeBanner
        isLoading={isLoading}
        subscription={activeSubscription}
        isAllowed={canChangePlan}
      />

      <Flex className={styles.plansRow}>
        <div className={styles.featureLabelSpacer} />
        {groupedPlans.map(plan => (
          <PlanCard
            key={plan.slug}
            plan={plan}
            currentPlan={currentPlanPricing}
            selectedInterval={selectedIntervals[plan.slug]}
            onIntervalChange={interval =>
              setIntervalOverrides(prev => ({
                ...prev,
                [plan.slug]: interval
              }))
            }
            allowAction={canChangePlan}
            onConfirmPlanChange={payload =>
              confirmPlanChangeHandle.openWithPayload(payload)
            }
          />
        ))}
      </Flex>

      <FeatureTable
        features={sortedFeatures}
        plans={groupedPlans}
        selectedIntervals={selectedIntervals}
        isLoading={isFeaturesLoading}
      />

      <ConfirmPlanChangeDialog handle={confirmPlanChangeHandle} />
    </ViewContainer>
  );
}
