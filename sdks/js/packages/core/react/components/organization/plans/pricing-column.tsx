import { useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as _ from 'lodash';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Skeleton,
  Image,
  toast,
  Text,
  Flex,
  Tabs,
  Amount
} from '@raystack/apsara';
import {
  IntervalKeys,
  IntervalLabelMap,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import { usePlans } from './hooks/usePlans';
import { PlanChangeAction, getPlanChangeAction } from '~/react/utils';
import {
  DEFAULT_DATE_FORMAT,
  DEFAULT_DATE_SHORT_FORMAT,
  SUBSCRIPTION_STATES
} from '~/react/utils/constants';
import checkCircle from '~/react/assets/check-circle.svg';

import plansStyles from './plans.module.css';
import { timestampToDayjs } from '~/utils/timestamp';

interface PricingColumnHeaderProps {
  plan: PlanIntervalPricing;
  selectedInterval: IntervalKeys;
}

const PricingColumnHeader = ({
  plan,
  selectedInterval
}: PricingColumnHeaderProps) => {
  const { config } = useFrontier();
  const selectedIntervalPricing = plan.intervals[selectedInterval];
  const showPerMonthPrice =
    config?.billing?.showPerMonthPrice === true && selectedInterval === 'year';

  const perIntervalLabel = showPerMonthPrice
    ? 'per seat/month'
    : `per seat/${selectedInterval}`;

  const amount = showPerMonthPrice
    ? (selectedIntervalPricing?.amount || 0) / 12
    : selectedIntervalPricing?.amount;

  const actualPerMonthAmount = plan.intervals['month']?.amount || 0;
  const discount =
    showPerMonthPrice && actualPerMonthAmount > 0
      ? ((actualPerMonthAmount - amount) * 100) / actualPerMonthAmount
      : 0;

  const showDiscount = showPerMonthPrice && discount > 0;
  const discountText = showDiscount ? (discount * -1).toFixed(0) + '%' : '';

  return (
    <Flex gap={3} direction="column">
      <Flex align="center" gap={3}>
        <Text size="regular" weight="medium" className={plansStyles.planTitle}>
          {plan.title}
        </Text>
        {showDiscount ? (
          <Flex className={plansStyles.discountText}>
            <Text weight="medium" variant="tertiary">
              {discountText}
            </Text>
          </Flex>
        ) : null}
      </Flex>
      <Flex gap={2} align="end">
        <Text className={plansStyles.planPrice}>
          <Amount
            value={amount}
            currency={selectedIntervalPricing?.currency}
            hideDecimals={config?.billing?.hideDecimals}
          />
        </Text>
        <Text size="small" variant="tertiary">
          {perIntervalLabel}
        </Text>
      </Flex>
      <Text size="small" variant="tertiary">
        {plan?.description}
      </Text>
    </Flex>
  );
};

interface FeaturesListProps {
  features: string[];
  plan?: IntervalPricingWithPlan;
}

const FeaturesList = ({ features, plan }: FeaturesListProps) => {
  return features.map(feature => {
    const planFeature = _.get(plan?.features, feature, {
      metadata: {}
    });
    const productMetaDataFeatureValues = plan?.productNames
      .map(name => _.get(planFeature.metadata, name))
      .filter(value => value !== undefined);
    // picking the first value for feature metadata, in case of multiple products in a plan, there can be multiple metadata values.
    const value = productMetaDataFeatureValues?.[0] || '-';
    const isAvailable = value?.toLowerCase() === 'true';
    return (
      <Flex
        key={feature + '-' + plan?.planId}
        align="center"
        justify="start"
        className={plansStyles.featureCell}
      >
        {isAvailable ? (
          <Image src={checkCircle as unknown as string} alt="checked" />
        ) : value ? (
          <Text>{value}</Text>
        ) : (
          <Text>-</Text>
        )}
      </Flex>
    );
  });
};

interface PlanIntervalsProps {
  planSlug: string;
  planIntervals: IntervalKeys[];
  selectedInterval: IntervalKeys;
  // eslint-disable-next-line no-unused-vars
  onIntervalChange: (i: IntervalKeys) => void;
}

const PlanIntervals = ({
  planSlug,
  planIntervals,
  selectedInterval,
  onIntervalChange
}: PlanIntervalsProps) => {
  return planIntervals.length > 1 ? (
    <Tabs
      className={plansStyles.plansIntervalList}
      value={selectedInterval}
      onValueChange={value => onIntervalChange(value as IntervalKeys)}
    >
      <Tabs.List>
        {planIntervals.map(key => (
          <Tabs.Trigger
            value={key}
            key={key}
            className={plansStyles.plansIntervalListItem}
            data-test-id={`frontier-sdk-plan-interval-toggle-${planSlug}-${key}`}
          >
            <Text className={plansStyles.plansIntervalListItemText}>
              {IntervalLabelMap[key]}
            </Text>
          </Tabs.Trigger>
        ))}
      </Tabs.List>
    </Tabs>
  ) : null;
};

interface TrialLinkProps {
  planIds: string[];
  isUpgrade: boolean;
  planHasTrial: boolean;
  onButtonClick: () => void;
  disabled: boolean;
  dateFormat: string;
  'data-test-id'?: string;
}

const TrialLink = function TrialLink({
  disabled,
  planIds,
  isUpgrade,
  planHasTrial,
  dateFormat,
  onButtonClick = () => {},
  'data-test-id': dataTestId
}: TrialLinkProps) {
  const {
    isTrialCheckLoading,
    hasAlreadyTrialed,
    checkAlreadyTrialed,
    subscriptions,
    isCurrentlyTrialing
  } = usePlans();

  useEffect(() => {
    if (planHasTrial) {
      checkAlreadyTrialed(planIds);
    }
  }, [checkAlreadyTrialed, planHasTrial, planIds]);

  const trialSubscription = subscriptions.find(
    sub =>
      planIds.includes(sub.planId || '') &&
      sub.state === SUBSCRIPTION_STATES.TRIALING
  );

  const trialEndDate = trialSubscription?.trialEndsAt
    ? timestampToDayjs(trialSubscription?.trialEndsAt)?.format(dateFormat)
    : '';

  const showButton =
    isUpgrade && !hasAlreadyTrialed && planHasTrial && !isCurrentlyTrialing;
  return (
    <Flex
      className={plansStyles.trialWrapper}
      justify={'center'}
      align={'center'}
    >
      {isTrialCheckLoading ? (
        <Skeleton containerClassName={plansStyles.flex1} />
      ) : trialEndDate ? (
        <Text>Trial ends on: {trialEndDate}</Text>
      ) : showButton ? (
        <Button
          className={plansStyles.trialButton}
          variant="outline"
          color="neutral"
          size="small"
          onClick={onButtonClick}
          disabled={disabled}
          data-test-id={dataTestId}
        >
          <Text>Start a free trial</Text>
        </Button>
      ) : null}
    </Flex>
  );
};

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
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;
  const shortDateFormat = config?.shortDateFormat || DEFAULT_DATE_SHORT_FORMAT;
  const [isTrialCheckoutLoading, setIsTrialCheckoutLoading] = useState(false);
  const plans = useMemo(() => Object.values(plan.intervals), [plan.intervals]);

  const navigate = useNavigate({ from: '/plans' });

  const {
    checkoutPlan,
    isLoading,
    changePlan,
    verifyPlanChange,
    checkBasePlan
  } = usePlans();

  const planIntervals =
    plans.sort((a, b) => a.weightage - b.weightage).map(i => i.interval) || [];

  const [selectedInterval, setSelectedInterval] = useState<IntervalKeys>(() => {
    const activePlan = plans.find(p => p.planId === currentPlan?.planId);
    return activePlan?.interval || planIntervals[0] || 'year';
  });

  const onIntervalChange = (value: IntervalKeys) => {
    if (value) {
      setSelectedInterval(value);
    }
  };

  const selectedIntervalPricing = plan.intervals[selectedInterval];

  const action: PlanChangeAction = useMemo(() => {
    const isCurrentPlanSelectedPlan =
      selectedIntervalPricing?.planId === currentPlan?.planId;
    const isCurrentPlanBasePlan =
      checkBasePlan(selectedIntervalPricing?.planId) &&
      currentPlan?.planId === undefined;

    if (isCurrentPlanSelectedPlan || isCurrentPlanBasePlan) {
      return {
        disabled: true,
        btnLabel: 'Current Plan',
        btnLoadingLabel: 'Current Plan',
        btnVariant: 'outline',
        btnColor: 'neutral',
        btnSize: 'small',
        btnDoneLabel: ''
      };
    }

    const planAction = getPlanChangeAction(
      selectedIntervalPricing?.weightage,
      currentPlan?.weightage
    );
    return {
      disabled: false,
      ...planAction
    };
  }, [
    checkBasePlan,
    currentPlan?.planId,
    currentPlan?.weightage,
    selectedIntervalPricing?.planId,
    selectedIntervalPricing?.weightage
  ]);

  const isAlreadySubscribed = !_.isEmpty(currentPlan);
  const isUpgrade = action.btnLabel === 'Upgrade';

  const isCheckoutRequired =
    _.isEmpty(paymentMethod) && selectedIntervalPricing?.amount > 0;

  const planHasTrial = useMemo(
    () => plans.some(p => Number(p.trial_days) > 0),
    [plans]
  );
  const planIds = useMemo(() => plans.map(p => p.planId), [plans]);

  const onPlanActionClick = useCallback(() => {
    if (action?.showModal && !isCheckoutRequired && isAlreadySubscribed) {
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
            const changeDate = timestampToDayjs(planPhase?.effectiveAt)?.format(
              dateFormat
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
        isTrial: false,
        onSuccess: data => {
          window.location.href = data?.checkoutUrl as string;
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
    dateFormat,
    checkoutPlan
  ]);

  const checkoutTrial = () => {
    setIsTrialCheckoutLoading(true);
    checkoutPlan({
      planId: selectedIntervalPricing?.planId,
      isTrial: true,
      onSuccess: data => {
        setIsTrialCheckoutLoading(false);
        window.location.href = data?.checkoutUrl as string;
      }
    });
  };

  return (
    <Flex direction={'column'} style={{ flex: 1 }}>
      <Flex className={plansStyles.planInfoColumn} direction="column">
        <PricingColumnHeader plan={plan} selectedInterval={selectedInterval} />
        <Flex direction="column" gap={5}>
          {allowAction ? (
            <Button
              variant={action.btnVariant}
              color={action.btnColor}
              size={action.btnSize}
              className={plansStyles.planActionBtn}
              onClick={onPlanActionClick}
              disabled={action?.disabled || isLoading}
              loading={isLoading && !isTrialCheckoutLoading}
              loaderText={`${action.btnLoadingLabel}...`}
              data-test-id={`frontier-sdk-plan-action-button-${plan?.slug}`}
            >
              {action.btnLabel}
            </Button>
          ) : null}
          <PlanIntervals
            planIntervals={planIntervals}
            selectedInterval={selectedInterval}
            onIntervalChange={onIntervalChange}
            planSlug={plan?.slug}
          />
          {allowAction ? (
            <TrialLink
              data-test-id={`frontier-sdk-plan-trial-link-${plan?.slug}`}
              planIds={planIds}
              isUpgrade={isUpgrade}
              planHasTrial={planHasTrial}
              onButtonClick={checkoutTrial}
              disabled={action?.disabled || isLoading}
              dateFormat={shortDateFormat}
            />
          ) : null}
        </Flex>
      </Flex>
      <Flex direction="column">
        <Flex
          align="center"
          justify="start"
          className={plansStyles.featureCell}
        >
          <Text size="small" className={plansStyles.featureTableHeading}>
            {plan.title}
          </Text>
        </Flex>
      </Flex>
      <FeaturesList features={features} plan={selectedIntervalPricing} />
    </Flex>
  );
};
