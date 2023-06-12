import { styled } from "@raystack/apsara";
import type { NumberSize, ResizableProps, ResizeDirection } from "re-resizable";
import { Resizable } from "re-resizable";
import { useState } from "react";

interface ResizablePropsInterface extends ResizableProps {
  children?: React.ReactNode;
}

const ResizableComponent = styled(Resizable, {
  display: "flex",
  borderRight: "solid 2px #ddd",
});

const ResizableContainer = ({
  children,
  ...props
}: ResizablePropsInterface) => {
  const [width, setWidth] = useState(Number(props.defaultSize?.width));
  const [height, _] = useState(props.defaultSize?.height);
  return (
    <ResizableComponent
      size={{ width, height }}
      enable={{
        top: false,
        right: false,
        bottom: false,
        left: false,
        topRight: false,
        bottomRight: false,
        bottomLeft: false,
        topLeft: false,
      }}
      maxHeight="100vh"
      onResizeStop={(
        event: MouseEvent | TouchEvent,
        direction: ResizeDirection,
        elementRef: HTMLElement,
        delta: NumberSize
      ) => {
        setWidth(width + delta.width);
      }}
      {...props}
    >
      {children}
    </ResizableComponent>
  );
};

export default ResizableContainer;
