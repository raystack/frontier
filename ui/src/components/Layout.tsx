import { Box, Flex } from "@raystack/apsara";
import React from "react";

type Props = {
  header?: React.ReactNode;
  sidebar?: React.ReactNode;
  sidebarRight?: React.ReactNode;
  children?: React.ReactNode;
};

const RESIZABLE = {
  width: 220,
  height: "100%",
  minWidth: "220px",
  maxWidth: "330px",
  minHeight: "100%",
};

export default function Layout({ header, children, sidebar }: Props) {
  return (
    <Box>
      <Flex direction={"row"} style={containerStyle}>
        <Flex direction="column" style={{ width: "250px" }}>
          {sidebar}
        </Flex>
        <Flex direction="column" style={{ flexGrow: 1, position: "relative" }}>
          {header}
          <Flex style={contentContainerStyle}>{children}</Flex>
        </Flex>
      </Flex>
    </Box>
  );
}

const containerStyle = {
  width: "100vw",
  height: "100vh",
  minHeight: "100vh",
  overflow: "hidden",
  alignItems: "stretch",
};



const contentContainerStyle = {
  overflow: "auto",
  position: "relative",
  flexGrow: 1,
};
