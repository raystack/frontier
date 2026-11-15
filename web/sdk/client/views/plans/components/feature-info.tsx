'use client';

import { Fragment } from 'react';
import { Link, Popover, Text } from '@raystack/apsara';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { parseFeatureInfo } from '~/client/views/plans/helpers/feature-rows';
import styles from './feature-table.module.css';

interface FeatureInfoProps {
  info: string;
}

export function FeatureInfo({ info }: FeatureInfoProps) {
  return (
    <Text size="small">
      {parseFeatureInfo(info).map((part, index) =>
        part.type === 'link' ? (
          <Link
            key={index}
            href={part.url}
            external={!part.url.toLowerCase().startsWith('mailto:')}
            size="small"
            data-test-id="frontier-sdk-plans-feature-info-link"
          >
            {part.label}
          </Link>
        ) : (
          <Fragment key={index}>{part.text}</Fragment>
        )
      )}
    </Text>
  );
}

interface FeatureInfoPopoverProps {
  title: string;
  info: string;
}

export function FeatureInfoPopover({ title, info }: FeatureInfoPopoverProps) {
  return (
    <Popover>
      <Popover.Trigger
        openOnHover
        className={styles.infoTrigger}
        aria-label={`More information about ${title}`}
        data-test-id="frontier-sdk-plans-feature-info-btn"
      >
        <InfoCircledIcon width={16} height={16} />
      </Popover.Trigger>
      <Popover.Content
        side="top"
        align="start"
        className={styles.infoContent}
        render={<div aria-label={`About ${title}`} />}
      >
        <FeatureInfo info={info} />
      </Popover.Content>
    </Popover>
  );
}
