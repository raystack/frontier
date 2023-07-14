import { Text } from "@raystack/apsara";
import React, { ComponentPropsWithRef, useCallback } from "react";
import { useShield } from "../contexts/ShieldContext";
import { MagicLink } from "./magiclink";
import { OIDCButton } from "./oidc";

const styles = {
  container: {
    fontSize: "12px",

    width: "28rem",
    maxWidth: "100%",
    padding: "1.5rem",

    color: "rgb(60, 74, 90)",
    backgroundColor: "#FFF",

    display: "flex",
    flexDirection: "column",
    alignItems: "center",
    gap: "8px",
  },
  logoContainer: {
    marginBottom: "1.5rem",
  },
  titleContainer: {
    fontSize: "14px",
    fontWeight: "bold",
    marginBottom: "1rem",
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

const defaultLogo = (
  <img
    src="logo.svg"
    style={{ borderRadius: "8px", width: "40px", height: "40px" }}
  ></img>
);

type SignedInProps = ComponentPropsWithRef<typeof Container>;
export const SignedIn = (props: SignedInProps) => {
  const { client, strategies = [] } = useShield();

  const clickHandler = useCallback(
    async (name: string) => {
      const {
        data: { endpoint },
      } = await client.getAuthStrategyEndpoint(name);
      window.location.href = endpoint;
    },
    [strategies]
  );

  const mailotp = strategies.find((s) => s.name === "mailotp");
  const filteredOIDC = strategies.filter((s) => s.name !== "mailotp");

  return (
    <Container {...props}>
      {mailotp && <MagicLink />}
      {mailotp && (
        <fieldset style={{ ...styles.fieldset, boxSizing: "border-box" }}>
          <legend style={styles.legend}>or</legend>
        </fieldset>
      )}
      {filteredOIDC.map((s, index) => {
        return (
          <OIDCButton key={index} onClick={() => clickHandler(s.name)}>
            {s.name}
          </OIDCButton>
        );
      })}
    </Container>
  );
};

type ContainerProps = ComponentPropsWithRef<"div"> & {
  children?: React.ReactNode;
  shadow?: "none" | "xs" | "sm" | "md" | "lg";
  radius?: "none" | "xs" | "sm" | "md" | "lg";
  title?: string;
  logo?: React.ReactNode;
};

export const Container = ({
  children,
  shadow = "sm",
  radius = "md",
  title = "Sign in",
  logo,
}: ContainerProps) => (
  <div
    style={{
      ...styles.container,
      flexDirection: "column",
      boxShadow: shadowOptions[shadow],
      borderRadius: borderRadiusOptions[radius],
    }}
  >
    <div style={styles.logoContainer}>{logo ? logo : defaultLogo}</div>
    <div style={styles.titleContainer}>
      <Text size={6}>{title}</Text>
    </div>
    {children}
  </div>
);
