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

  if (!active) return children;

  return (
    <Tooltip>
      <Tooltip.Trigger render={<span className={styles.trigger} />}>
        {children}
      </Tooltip.Trigger>
      <Tooltip.Content
        side="right"
        align="start"
        sideOffset={-40}
        alignOffset={-10}
      >
        {message}
      </Tooltip.Content>
    </Tooltip>
  );
};
