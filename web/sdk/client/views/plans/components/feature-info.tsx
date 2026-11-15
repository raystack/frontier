'use client';

import { Fragment, useId } from 'react';
import { Link, Popover } from '@raystack/apsara';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import { parseFeatureInfo } from '~/client/views/plans/helpers/feature-rows';
import styles from './feature-table.module.css';

interface FeatureInfoProps {
  id: string;
  info: string;
}

function FeatureInfo({ id, info }: FeatureInfoProps) {
  return (
    <span id={id}>
      {parseFeatureInfo(info).map((part, index) =>
        part.type === 'link' ? (
          <Link
            key={index}
            href={part.url}
            external={!part.url.startsWith('mailto:')}
            underline
            data-test-id="frontier-sdk-plans-feature-info-link"
          >
            {part.label}
          </Link>
        ) : (
          <Fragment key={index}>{part.text}</Fragment>
        )
      )}
    </span>
  );
}

interface FeatureInfoPopoverProps {
  title: string;
  info: string;
}

export function FeatureInfoPopover({ title, info }: FeatureInfoPopoverProps) {
  const descriptionId = useId();
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
        render={
          <div aria-label={`About ${title}`} aria-describedby={descriptionId} />
        }
      >
        <FeatureInfo id={descriptionId} info={info} />
      </Popover.Content>
    </Popover>
  );
}
