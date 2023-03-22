import React from "react";
import Resizable from "~/components/Resizable";
import { BaseContainer } from "~/primitives/container/BaseContainer";
import { Flex } from "~/primitives/flex/Flex";
import { styled } from "~/stitches";

type Props = {
  header?: React.ReactNode;
  sidebar?: React.ReactNode;
  sidebarRight?: React.ReactNode;
  children?: React.ReactNode;
};

const RESIZABLE = {
  width: 220,
  height: "100%",
  minWidth: "260px",
  maxWidth: "330px",
  minHeight: "100%",
};
const ResizableContainer = styled(Resizable, {});
export default function Layout({ header, children, sidebar }: Props) {
  return (
    <BaseContainer>
      <Flex direction={"row"} css={containerStyle}>
        <ResizableContainer
          minWidth={RESIZABLE.minWidth}
          maxWidth={RESIZABLE.maxWidth}
          minHeight={RESIZABLE.minHeight}
          defaultSize={{
            width: RESIZABLE.width,
            height: RESIZABLE.height,
          }}
        >
          {sidebar}
        </ResizableContainer>
        <Flex as={"main"} direction="column" css={mainContainerStyle}>
          {header}
          <Flex css={contentContainerStyle}>{children}</Flex>
        </Flex>
      </Flex>
    </BaseContainer>
  );
}

const containerStyle = {
  width: "100vw",
  height: "100vh",
  minHeight: "100vh",
  overflow: "hidden",
  alignItems: "stretch",
};

const mainContainerStyle = {
  flexGrow: 1,
  position: "relative",
};

const contentContainerStyle = {
  overflow: "auto",
  position: "relative",
  flexGrow: 1,
};
