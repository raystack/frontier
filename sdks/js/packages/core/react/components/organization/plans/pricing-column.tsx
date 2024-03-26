import { useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import dayjs from 'dayjs';
import { toast } from 'sonner';
import * as _ from 'lodash';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { Button, Flex, Text, ToggleGroup, Image } from '@raystack/apsara';
import {
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import { usePlans } from './hooks/usePlans';
import { PlanChangeAction, getPlanChangeAction } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import checkCircle from '~/react/assets/check-circle.svg';
import Amount from '~/react/components/helpers/Amount';

import plansStyles from './plans.module.css';

interface PlanPricingColumnProps {
  plan: PlanIntervalPricing;
  features: string[];
  currentPlan?: IntervalPricingWithPlan;
  allowAction: boolean;
}

export const PlanPricingColumn = ({
  plan,
  features,
  currentPlan,
  allowAction
}: PlanPricingColumnProps) => {
  const { config, paymentMethod } = useFrontier();

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
        btnVariant: 'secondary',
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
  const isCheckoutRequired =
    _.isEmpty(paymentMethod) && selectedIntervalPricing.amount > 0;

  const onPlanActionClick = useCallback(() => {
    if (action?.showModal && !isCheckoutRequired) {
      navigate({
        to: '/plans/confirm-change/$planId',
        params: {
          planId: selectedIntervalPricing?.planId
        }
      });
    } else if (isAlreadySubscribed && !isCheckoutRequired) {
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
    isCheckoutRequired,
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
              variant={action.btnVariant}
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
        const planFeature = _.get(selectedIntervalPricing.features, feature, {
          metadata: {}
        });
        const productMetaDataFeatureValues =
          selectedIntervalPricing.productNames
            .map(name => _.get(planFeature.metadata, name))
            .filter(value => value !== undefined);
        // piciking the first value for feature metadata, in case of muliple products in a plan, there can be multiple metadata values.
        const value = productMetaDataFeatureValues[0];
        const isAvailable = value?.toLowerCase() === 'true';
        return (
          <Flex
            key={feature + '-' + plan?.slug}
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
            ) : (
              <Text>-</Text>
            )}
          </Flex>
        );
      })}
    </Flex>
  );
};
