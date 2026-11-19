'use client';

import { IconButton, Link, Popover } from '@raystack/apsara';
import { InfoCircledIcon } from '@radix-ui/react-icons';
import styles from './feature-table.module.css';

/*
 * Matches [label](url) for http, https and mailto urls only.
 * - Label needs a non-space character, so [ ](url) stays text.
 * - Label can't hold brackets, so [[a](url) links only "a".
 * - http(s) needs // and a host, so https:x and https://#a stay text.
 * - Url may hold balanced parentheses one level deep, e.g. .../Foo_(bar).
 */
const LINK_PATTERN =
  /\[(\s*[^[\]\s][^[\]]*)\]\(((?:https?:\/\/(?=[^\s()/?#])|mailto:)(?:[^()\s]|\([^()\s]*\))+)\)/;

interface FeatureInfoProps {
  title: string;
  info: string;
}

export function FeatureInfo({ title, info }: FeatureInfoProps) {
  return (
    <Popover>
      <Popover.Trigger
        openOnHover
        render={
          <IconButton
            size={2}
            aria-label={`About ${title}`}
            className={styles.infoTrigger}
            data-test-id="frontier-sdk-plans-feature-info-btn"
          />
        }
      >
        <InfoCircledIcon />
      </Popover.Trigger>
      <Popover.Content
        side="top"
        render={<div aria-label={`About ${title}`} />}
      >
        {/* split() returns [text, label, url, text, label, url, ..., text] */}
        {info.split(LINK_PATTERN).map((part, i, parts) => {
          if (i % 3 === 0) return part;
          if (i % 3 === 2) return null;
          const url = parts[i + 1];
          return (
            <Link
              key={i}
              href={url}
              external={url.startsWith('http')}
              underline
              data-test-id="frontier-sdk-plans-feature-info-link"
            >
              {part}
            </Link>
          );
        })}
      </Popover.Content>
    </Popover>
  );
}
