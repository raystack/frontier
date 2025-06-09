import { Flex, Image, Dialog } from '@raystack/apsara/v1';
import React, { useState } from 'react';
import closeClose from '~/react/assets/close-close.svg';
import closeDefault from '~/react/assets/close-default.svg';
import resizeCollapse from '~/react/assets/resize-collapse.svg';
import resizeDefault from '~/react/assets/resize-default.svg';
import resizeExpand from '~/react/assets/resize-expand.svg';
// @ts-ignore
import styles from './window.module.css';

interface WindowProps extends React.HTMLAttributes<HTMLDialogElement> {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children?: React.ReactNode;
}

export const Window = ({
  open = false,
  onOpenChange,
  children,
  ...props
}: WindowProps) => {
  const [zoom, setZoom] = useState(false);
  const [isCloseActive, setCloseActive] = useState(false);
  const [isZoomActive, setZoomActive] = useState(false);
  return (
    <Dialog open={open} onOpenChange={onOpenChange} {...props}>
      <Dialog.Content
        className={`${styles.container} ${
          zoom ? styles.dialogContentZoomin : styles.dialogContentZoomout
        }`}
        overlayClassName={styles.overlay}
      >
        <div style={{ position: 'absolute', inset: 0 }}>{children}</div>
        <div
          style={{
            position: 'absolute',
            top: 0,
            padding: '16px'
          }}
        >
          <Flex gap={3}>
            <Image
              onMouseOver={() => setCloseActive(true)}
              onMouseOut={() => setCloseActive(false)}
              alt="close-button"
              src={
                isCloseActive
                  ? (closeClose as unknown as string)
                  : (closeDefault as unknown as string)
              }
              onClick={() => onOpenChange && onOpenChange(false)}
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-window-close-button"
            />
            <Image
              onMouseOver={() => setZoomActive(true)}
              onMouseOut={() => setZoomActive(false)}
              alt="maximize-toggle-button"
              src={
                isZoomActive
                  ? zoom
                    ? (resizeCollapse as unknown as string)
                    : (resizeExpand as unknown as string)
                  : (resizeDefault as unknown as string)
              }
              onClick={() => setZoom(!zoom)}
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-window-maximize-button"
            />
          </Flex>
        </div>
      </Dialog.Content>
    </Dialog>
  );
};
