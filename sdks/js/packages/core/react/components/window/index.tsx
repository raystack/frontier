import { Dialog, Flex, Image } from '@raystack/apsara';
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
      {/* @ts-ignore */}
      <Dialog.Content
        className={`${styles.container} ${
          zoom ? styles.dialogContentZoomin : styles.dialogContentZoomout
        }`}
        overlayClassname={styles.overlay}
      >
        <div style={{ position: 'absolute', inset: 0 }}>{children}</div>
        <div
          style={{
            position: 'absolute',
            top: 0,
            padding: '16px'
          }}
        >
          <Flex gap="small">
            <Image
              onMouseOver={() => setCloseActive(true)}
              onMouseOut={() => setCloseActive(false)}
              alt="close-button"
              // @ts-ignore
              src={isCloseActive ? closeClose : closeDefault}
              onClick={() => onOpenChange && onOpenChange(false)}
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-window-close-button"
            />
            <Image
              onMouseOver={() => setZoomActive(true)}
              onMouseOut={() => setZoomActive(false)}
              alt="maximize-toggle-button"
              // @ts-ignore
              src={
                isZoomActive
                  ? zoom
                    ? resizeCollapse
                    : resizeExpand
                  : resizeDefault
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
