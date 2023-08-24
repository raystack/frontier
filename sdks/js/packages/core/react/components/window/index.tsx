import { Dialog, Flex, Image } from '@raystack/apsara';
import { useState } from 'react';
import closeClose from '~/react/assets/close-close.svg';
import closeDefault from '~/react/assets/close-default.svg';
import resizeCollapse from '~/react/assets/resize-collapse.svg';
import resizeDefault from '~/react/assets/resize-default.svg';
import resizeExpand from '~/react/assets/resize-expand.svg';

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
        style={{
          padding: 0,
          ...(zoom
            ? { width: '100vw', height: '100vh', maxHeight: 'reset' }
            : { width: '80vw', height: '80vh' })
        }}
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
              alt="cross"
              // @ts-ignore
              src={isCloseActive ? closeClose : closeDefault}
              onClick={() => onOpenChange && onOpenChange(false)}
            />
            <Image
              onMouseOver={() => setZoomActive(true)}
              onMouseOut={() => setZoomActive(false)}
              alt="cross"
              // @ts-ignore
              src={
                isZoomActive
                  ? zoom
                    ? resizeCollapse
                    : resizeExpand
                  : resizeDefault
              }
              onClick={() => setZoom(!zoom)}
            />
          </Flex>
        </div>
      </Dialog.Content>
    </Dialog>
  );
};
