import { useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import dayjs from 'dayjs';
import { toast } from 'sonner';
import * as _ from 'lodash';
import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import { Button, Flex, Text, ToggleGroup, Image } from '@raystack/apsara';
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
  DEFAULT_DATE_SHORT_FORMAT
} from '~/react/utils/constants';
import checkCircle from '~/react/assets/check-circle.svg';
import Amount from '~/react/components/helpers/Amount';

import plansStyles from './plans.module.css';
import Skeleton from 'react-loading-skeleton';

interface FeaturesListProps {
  features: string[];
  plan: IntervalPricingWithPlan;
}

const FeaturesList = ({ features, plan }: FeaturesListProps) => {
  return features.map(feature => {
    const planFeature = _.get(plan.features, feature, {
      metadata: {}
    });
    const productMetaDataFeatureValues = plan.productNames
      .map(name => _.get(planFeature.metadata, name))
      .filter(value => value !== undefined);
    // piciking the first value for feature metadata, in case of muliple products in a plan, there can be multiple metadata values.
    const value = productMetaDataFeatureValues[0];
    const isAvailable = value?.toLowerCase() === 'true';
    return (
      <Flex
        key={feature + '-' + plan.planId}
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
  });
};

interface PlanIntervalsProps {
  planIntervals: IntervalKeys[];
  selectedInterval: IntervalKeys;
  onIntervalChange: (i: IntervalKeys) => void;
}

const PlanIntervals = ({
  planIntervals,
  selectedInterval,
  onIntervalChange
}: PlanIntervalsProps) => {
  return planIntervals.length > 1 ? (
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
  ) : null;
};

interface TrialLinkProps {
  planIds: string[];
  isUpgrade: boolean;
  planHasTrial: boolean;
  onButtonClick: () => void;
  disabled: boolean;
  dateFormat: string;
}

const TrialLink = function TrialLink({
  disabled,
  planIds,
  isUpgrade,
  planHasTrial,
  dateFormat,
  onButtonClick = () => {}
}: TrialLinkProps) {
  const {
    isTrailCheckLoading,
    hasAlreadyTrailed,
    checkAlreadyTrialed,
    trailSubscription
  } = usePlans();

  useEffect(() => {
    if (planHasTrial) {
      checkAlreadyTrialed(planIds);
    }
  }, [checkAlreadyTrialed, planHasTrial, planIds]);

  const trailEndDate = planIds.includes(trailSubscription?.plan_id || '')
    ? dayjs(trailSubscription?.trial_ends_at).format(dateFormat)
    : '';

  const showButton = isUpgrade && !hasAlreadyTrailed && planHasTrial;
  return (
    <Flex
      className={plansStyles.trialWrapper}
      justify={'center'}
      align={'center'}
    >
      {isTrailCheckLoading ? (
        <Skeleton containerClassName={plansStyles.flex1} />
      ) : trailEndDate ? (
        <Text>Trial ends on: {trailEndDate}</Text>
      ) : showButton ? (
        <Button
          className={plansStyles.trialButton}
          variant={'secondary'}
          onClick={onButtonClick}
          disabled={disabled}
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
  const [isTrailCheckoutLoading, setIsTrialCheckoutLoading] = useState(false);
  const plans = useMemo(() => Object.values(plan.intervals), [plan.intervals]);

  const navigate = useNavigate({ from: '/plans' });

  const { checkoutPlan, isLoading, changePlan, verifyPlanChange } = usePlans();

  const planIntervals =
    plans.sort((a, b) => a.weightage - b.weightage).map(i => i.interval) || [];

  const [selectedInterval, setSelectedInterval] = useState<IntervalKeys>(() => {
    const activePlan = plans.find(p => p.planId === currentPlan?.planId);
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
  const isUpgrade = action.btnLabel === 'Upgrade';

  const isCheckoutRequired =
    _.isEmpty(paymentMethod) && selectedIntervalPricing.amount > 0;

  const planHasTrial = useMemo(
    () => plans.some(p => Number(p.trial_days) > 0),
    [plans]
  );
  const planIds = useMemo(() => plans.map(p => p.planId), [plans]);

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
        window.location.href = data?.checkout_url as string;
      }
    });
  };

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
              {isLoading && !isTrailCheckoutLoading
                ? `${action.btnLoadingLabel}....`
                : action.btnLabel}
            </Button>
          ) : null}
          <PlanIntervals
            planIntervals={planIntervals}
            selectedInterval={selectedInterval}
            onIntervalChange={onIntervalChange}
          />
          <TrialLink
            planIds={planIds}
            isUpgrade={isUpgrade}
            planHasTrial={planHasTrial}
            onButtonClick={checkoutTrial}
            disabled={action?.disabled || isLoading}
            dateFormat={shortDateFormat}
          />
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
      <FeaturesList features={features} plan={selectedIntervalPricing} />
    </Flex>
  );
};