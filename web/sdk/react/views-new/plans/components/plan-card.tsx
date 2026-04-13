'use client';

import { useCallback, useMemo } from 'react';
import {
  Amount,
  Button,
  Flex,
  Tabs,
  Text
} from '@raystack/apsara-v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePlans } from '~/react/views/plans/hooks/usePlans';
import { PlanChangeAction, getPlanChangeAction } from '~/react/utils';
import {
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import styles from './plan-card.module.css';

export interface PlanCardProps {
  plan: PlanIntervalPricing;
  currentPlan?: IntervalPricingWithPlan;
  selectedInterval: IntervalKeys;
  onIntervalChange: (interval: IntervalKeys) => void;
  allowAction: boolean;
  onConfirmPlanChange?: (payload: { planId: string; amount?: number; currency?: string }) => void;
}

export function PlanCard({
  plan,
  currentPlan,
  selectedInterval,
  onIntervalChange,
  allowAction,
  onConfirmPlanChange
}: PlanCardProps) {
  const { config } = useFrontier();

  const plans = useMemo(
    () => Object.values(plan.intervals),
    [plan.intervals]
  );

  const { checkBasePlan } = usePlans();

  const planIntervals = useMemo(
    () =>
      plans
        .sort((a, b) => a.weightage - b.weightage)
        .map(i => i.interval),
    [plans]
  );

  const selectedIntervalPricing = plan.intervals[selectedInterval];

  const showPerMonthPrice =
    config?.billing?.showPerMonthPrice === true && selectedInterval === 'year';

  const perIntervalLabel = showPerMonthPrice
    ? 'per seat/month'
    : `per seat/${selectedInterval}`;

  const amount = showPerMonthPrice
    ? (selectedIntervalPricing?.amount || 0) / 12
    : selectedIntervalPricing?.amount;

  const action: PlanChangeAction = useMemo(() => {
    const isCurrentPlanSelectedPlan =
      selectedIntervalPricing?.planId === currentPlan?.planId;
    const isCurrentPlanBasePlan =
      checkBasePlan(selectedIntervalPricing?.planId) &&
      currentPlan?.planId === undefined;

    if (isCurrentPlanSelectedPlan || isCurrentPlanBasePlan) {
      return {
        disabled: true,
        btnLabel: 'Current',
        btnLoadingLabel: 'Current',
        btnVariant: 'outline',
        btnColor: 'neutral',
        btnSize: 'small',
        btnDoneLabel: ''
      };
    }

    return {
      disabled: false,
      ...getPlanChangeAction(
        selectedIntervalPricing?.weightage,
        currentPlan?.weightage
      )
    };
  }, [
    checkBasePlan,
    currentPlan?.planId,
    currentPlan?.weightage,
    selectedIntervalPricing?.planId,
    selectedIntervalPricing?.weightage
  ]);

  const onPlanActionClick = useCallback(() => {
    onConfirmPlanChange?.({
      planId: selectedIntervalPricing?.planId,
      amount: selectedIntervalPricing?.amount,
      currency: selectedIntervalPricing?.currency
    });
  }, [onConfirmPlanChange, selectedIntervalPricing]);

  const isFree = !selectedIntervalPricing?.amount;

  return (
    <Flex direction="column" gap={6} className={styles.card}>
      <Flex direction="column" gap={3}>
        <Text size="small" weight="medium">
          {plan.title}
        </Text>
        <Flex gap={1} align="end">
          {isFree ? (
            <span className={styles.priceLabel}>Free</span>
          ) : (
            <>
              <Text className={styles.priceLabel}>
                <Amount
                  value={amount || 0}
                  currency={selectedIntervalPricing?.currency}
                  hideDecimals={config?.billing?.hideDecimals}
                  valueInMinorUnits={false}
                />
              </Text>
              <Text size="small" variant="secondary">
                {perIntervalLabel}
              </Text>
            </>
          )}
        </Flex>
      </Flex>

      <Flex direction="column" gap={5} align="end">
        {allowAction ? (
          <Button
            variant={action.btnVariant}
            color={action.btnColor}
            size={action.btnSize}
            className={styles.actionBtn}
            onClick={onPlanActionClick}
            disabled={action?.disabled}
            data-test-id={`frontier-sdk-plan-action-button-${plan?.slug}`}
          >
            {action.btnLabel}
          </Button>
        ) : null}
        {planIntervals.length > 1 ? (
          <Tabs
            value={selectedInterval}
            onValueChange={value => onIntervalChange(value as IntervalKeys)}
          >
            <Tabs.List>
              {planIntervals.map(key => (
                <Tabs.Tab
                  value={key}
                  key={key}
                  data-test-id={`frontier-sdk-plan-interval-toggle-${plan.slug}-${key}`}
                >
                  {IntervalLabelMap[key]}
                </Tabs.Tab>
              ))}
            </Tabs.List>
          </Tabs>
        ) : null}
      </Flex>
    </Flex>
  );
}
