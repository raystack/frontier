import { Flex, Link, Text } from "@raystack/apsara";
import React, { ComponentPropsWithRef, useCallback } from "react";
import { useFrontier } from "../contexts/FrontierContext";
import { Container } from "./Container";
import { Header } from "./Header";
import { MagicLink } from "./magiclink";
import { OIDCButton } from "./oidc";

const styles = {
  titleContainer: {
    fontWeight: "400",
  },
};

type SignedInProps = ComponentPropsWithRef<typeof Container> & {
  logo?: React.ReactNode;
  title?: string;
};
export const SignedIn = ({
  logo,
  title = "Login to Raypoint",
  ...props
}: SignedInProps) => {
  const { config } = useFrontier();
  const { client, strategies = [] } = useFrontier();

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
      <Header logo={logo} title={title} />
      <Flex direction="column" style={{ width: "100%", gap: "var(--pd-16)" }}>
        {filteredOIDC.map((s, index) => {
          return (
            <OIDCButton key={index} onClick={() => clickHandler(s.name)}>
              {s.name}
            </OIDCButton>
          );
        })}

        {mailotp && <MagicLink />}
      </Flex>
      <div style={styles.titleContainer}>
        <Text size={2}>
          Don’t have an account?{" "}
          <Link
            href={config.redirectSignup}
            style={{ color: "var(--foreground-accent)" }}
          >
            Signup
          </Link>
        </Text>
      </div>
    </Container>
  );
};
