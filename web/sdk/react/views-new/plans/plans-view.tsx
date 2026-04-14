'use client';

import { useEffect, useMemo, useState } from 'react';
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
    isAllPlansLoading
  } = useFrontier();
  console.log({ basePlan, allPlans });

  const { isFetching: isPermissionsFetching, isAllowed: canChangePlan } =
    useBillingPermission();

  const { data: featuresData } = useQuery(
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

  const isLoading =
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

  const [selectedIntervals, setSelectedIntervals] = useState<
    Record<string, IntervalKeys>
  >({});

  useEffect(() => {
    if (groupedPlans.length === 0) return;

    const defaults: Record<string, IntervalKeys> = {};
    groupedPlans.forEach(plan => {
      if (selectedIntervals[plan.slug]) return;
      const planIntervals = Object.values(plan.intervals)
        .sort((a, b) => a.weightage - b.weightage)
        .map(i => i.interval);
      const activePlanInterval = Object.values(plan.intervals).find(
        p => p.planId === activeSubscription?.planId
      );
      defaults[plan.slug] =
        activePlanInterval?.interval || planIntervals[0] || 'year';
    });

    if (Object.keys(defaults).length > 0) {
      setSelectedIntervals(prev => ({ ...prev, ...defaults }));
    }
  }, [groupedPlans, activeSubscription?.planId]);

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
            selectedInterval={
              selectedIntervals[plan.slug] ||
              Object.keys(plan.intervals)[0] as IntervalKeys
            }
            onIntervalChange={interval =>
              setSelectedIntervals(prev => ({
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
      />

      <ConfirmPlanChangeDialog handle={confirmPlanChangeHandle} />
    </ViewContainer>
  );
}
