import { Flex } from "@raystack/apsara";
import React, { ComponentPropsWithRef } from "react";

const styles = {
  container: {
    fontSize: "12px",
    width: "100%",
    minWidth: "220px",
    maxWidth: "480px",
    color: "var(--foreground-base)",

    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    gap: "32px",
  },
  logoContainer: {},
  titleContainer: {
    fontWeight: "400",
  },
  fieldset: {
    width: "100%",
    border: "1px solid transparent",
    borderTopColor: "rgb(205, 211, 223)",
    gridArea: "1 / 1",
    padding: 0,
    margin: "2px",
  },
  legend: {
    fontSize: "8px",
    margin: "auto",
    padding: "0 4px",
  },
};

const shadowOptions = {
  none: "none",
  xs: "0px 1px 2px 0px rgba(16, 24, 40, 0.06)",
  sm: "0px 1px 4px 0px rgba(0, 0, 0, 0.09)",
  md: "0px 4px 6px -2px rgba(16, 24, 40, 0.03), 0px 12px 16px -4px rgba(16, 24, 40, 0.08)",
  lg: "0px 8px 8px -4px rgba(16, 24, 40, 0.03), 0px 20px 24px -4px rgba(16, 24, 40, 0.08)",
};

const borderRadiusOptions = {
  none: "0",
  xs: "4px",
  sm: "8px",
  md: "16px",
  lg: "24px",
};

type ContainerProps = ComponentPropsWithRef<"div"> & {
  children?: React.ReactNode;
  shadow?: "none" | "xs" | "sm" | "md" | "lg";
  radius?: "none" | "xs" | "sm" | "md" | "lg";
};

export const Container = ({
  children,
  shadow = "none",
  radius = "md",
  style,
}: ContainerProps) => {
  return (
    <Flex
      style={{
        ...styles.container,
        ...style,
        flexDirection: "column",
        boxShadow: shadowOptions[shadow],
        borderRadius: borderRadiusOptions[radius],
      }}
    >
      {children}
    </Flex>
  );
};
