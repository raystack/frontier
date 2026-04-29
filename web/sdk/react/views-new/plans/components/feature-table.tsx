'use client';

import { Flex, Image, Skeleton, Text } from '@raystack/apsara-v1';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { get } from 'lodash';
import {
  IntervalKeys,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import checkCircle from '~/react/assets/check-circle.svg';
import styles from './feature-table.module.css';

interface FeatureValueCellProps {
  feature: string;
  plan?: IntervalPricingWithPlan;
}

function FeatureValueCell({ feature, plan }: FeatureValueCellProps) {
  const planFeature = get(plan?.features, feature, { metadata: {} });
  const productMetaDataFeatureValues = plan?.productNames
    .map(name => get(planFeature.metadata, name))
    .filter(value => value !== undefined);
  const value = productMetaDataFeatureValues?.[0] || '-';
  const isAvailable = value?.toLowerCase() === 'true';

  return (
    <Flex className={`${styles.cell} ${styles.valueCell}`}>
      {isAvailable ? (
        <Image
          src={checkCircle as unknown as string}
          alt="available"
          className={styles.checkIcon}
        />
      ) : (
        <Text size="regular" variant="secondary" style={{ textAlign: 'center', flex: 1 }}>
          {value}
        </Text>
      )}
    </Flex>
  );
}

export interface FeatureTableProps {
  features: string[];
  plans: PlanIntervalPricing[];
  selectedIntervals: Record<string, IntervalKeys>;
  isLoading?: boolean;
}

const SKELETON_ROW_COUNT = 5;

export function FeatureTable({
  features,
  plans,
  selectedIntervals,
  isLoading = false
}: FeatureTableProps) {
  if (isLoading) {
    return (
      <Flex className={styles.table}>
        <Flex direction="column" className={styles.featureColumn}>
          {Array.from({ length: SKELETON_ROW_COUNT }).map((_, i) => (
            <Flex key={i} className={styles.cell}>
              <Skeleton height={16} width="60%" />
            </Flex>
          ))}
        </Flex>
        {plans.map(plan => (
          <Flex
            key={plan.slug}
            direction="column"
            className={styles.planColumn}
          >
            {Array.from({ length: SKELETON_ROW_COUNT }).map((_, i) => (
              <Flex
                key={i}
                className={`${styles.cell} ${styles.valueCell}`}
              >
                <Skeleton height={16} width={40} />
              </Flex>
            ))}
          </Flex>
        ))}
      </Flex>
    );
  }

  if (features.length === 0) return null;

  return (
    <Flex className={styles.table}>
      <Flex direction="column" className={styles.featureColumn}>
        {features.map(feature => (
          <Flex key={feature} className={styles.cell}>
            <Text size="regular" className={styles.featureLabel}>
              {feature}
            </Text>
          </Flex>
        ))}
      </Flex>

      {plans.map(plan => {
        const interval = selectedIntervals[plan.slug];
        const planPricing = plan.intervals[interval];
        return (
          <Flex
            key={plan.slug}
            direction="column"
            className={styles.planColumn}
          >
            {features.map(feature => (
              <FeatureValueCell
                key={`${plan.slug}-${feature}`}
                feature={feature}
                plan={planPricing}
              />
            ))}
          </Flex>
        );
      })}
    </Flex>
  );
}
