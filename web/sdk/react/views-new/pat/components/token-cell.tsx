'use client';

import { Button, Flex, Text } from '@raystack/apsara-v1';
import styles from '../pat-view.module.css';

function KeyIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 16 16"
      fill="none"
    >
      <path
        d="M10.5 3.5C11.6046 3.5 12.5 4.39543 12.5 5.5M14.5 5.5C14.5 7.70914 12.7091 9.5 10.5 9.5C10.2662 9.5 10.037 9.47994 9.8142 9.44144C9.43885 9.37658 9.04134 9.45866 8.772 9.728L7 11.5H5.5V13H4V14.5H1.5V12.6213C1.5 12.2235 1.65804 11.842 1.93934 11.5607L6.272 7.228C6.54134 6.95866 6.62342 6.56115 6.55856 6.1858C6.52006 5.96297 6.5 5.73383 6.5 5.5C6.5 3.29086 8.29086 1.5 10.5 1.5C12.7091 1.5 14.5 3.29086 14.5 5.5Z"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export interface TokenCellProps {
  title: string;
  expiry: string;
  lastUsed: string;
  onClick?: () => void;
  onRevoke?: () => void;
}

export function TokenCell({ title, expiry, lastUsed, onClick, onRevoke }: TokenCellProps) {
  return (
    <Flex
      className={styles.tokenCell}
      gap={3}
      align="center"
      onClick={onClick}
      style={onClick ? { cursor: 'pointer' } : undefined}
    >
      <Flex style={{ flex: 1 }} gap={5} align="start">
        <div className={styles.tokenIcon}>
          <KeyIcon />
        </div>
        <Flex direction="column" gap={3} style={{ flex: 1, minWidth: 0 }}>
          <Text size="regular" weight="medium">
            {title}
          </Text>
          <Flex gap={3} align="center">
            {expiry && (
              <Text size="regular" variant="tertiary">
                {expiry}
              </Text>
            )}
            {expiry && lastUsed && (
              <Text size="regular" variant="tertiary">
                &bull;
              </Text>
            )}
            {lastUsed && (
              <Text size="regular" variant="tertiary">
                {lastUsed}
              </Text>
            )}
          </Flex>
        </Flex>
      </Flex>
      <Button
        variant="outline"
        color="neutral"
        size="small"
        onClick={e => {
          e.stopPropagation();
          onRevoke?.();
        }}
        data-test-id="frontier-sdk-revoke-pat-btn"
      >
        Revoke
      </Button>
    </Flex>
  );
}
