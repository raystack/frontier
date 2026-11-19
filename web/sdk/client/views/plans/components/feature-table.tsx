'use client';

import { Flex, Image, Skeleton, Text } from '@raystack/apsara';
import { get } from 'lodash';
import { Feature } from '@raystack/proton/frontier';
import {
  IntervalKeys,
  IntervalPricingWithPlan,
  PlanIntervalPricing
} from '~/src/types';
import checkCircle from '~/client/assets/check-circle.svg';
import { FeatureInfo } from '~/client/views/plans/components/feature-info';
import styles from './feature-table.module.css';

/*
 * Normalises a feature cell value to a string.
 * - Falsy values (missing, '', false, 0) become '-'.
 * - Plain values go through String(), so lowercasing a number or boolean doesn't throw.
 */
function normalizeFeatureValue(value: unknown): string {
  return value ? String(value) : '-';
}

/*
 * Reads the feature's info text.
 * - Only a non-blank string shows the info icon.
 */
function getFeatureInfo(feature: Feature): string {
  const info = feature.metadata?.info;
  return typeof info === 'string' ? info.trim() : '';
}

interface FeatureValueCellProps {
  feature: string;
  plan?: IntervalPricingWithPlan;
}

function FeatureValueCell({ feature, plan }: FeatureValueCellProps) {
  const planFeature = get(plan?.features, feature, { metadata: {} });
  const productMetaDataFeatureValues = plan?.productNames
    .map(name => get(planFeature.metadata, name))
    .filter(value => value !== undefined);
  const value = normalizeFeatureValue(productMetaDataFeatureValues?.[0]);
  const isAvailable = value.toLowerCase() === 'true';

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
  features: Feature[];
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
        {features.map(feature => {
          const info = getFeatureInfo(feature);
          return (
            <Flex key={feature.title} className={styles.cell}>
              <Text size="regular" className={styles.featureLabel}>
                {feature.title}
              </Text>
              {info && <FeatureInfo title={feature.title} info={info} />}
            </Flex>
          );
        })}
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
                key={`${plan.slug}-${feature.title}`}
                feature={feature.title}
                plan={planPricing}
              />
            ))}
          </Flex>
        );
      })}
    </Flex>
  );
}
