import { Tooltip } from '@raystack/apsara';
import { ReactNode } from 'react';
import styles from './disabled-tooltip.module.css';

export type DisabledTooltipProps = {
  disabled?: boolean;
  message?: string;
  children: ReactNode;
};

export const DisabledTooltip = ({
  disabled = false,
  message,
  children
}: DisabledTooltipProps) => {
  const active = disabled && !!message;

  return (
    <Tooltip>
      <Tooltip.Trigger
        disabled={!active}
        render={<span className={styles.trigger} />}
      >
        {children}
      </Tooltip.Trigger>
      {active && (
        <Tooltip.Content side="top" align="end">
          {message}
        </Tooltip.Content>
      )}
    </Tooltip>
  );
};
