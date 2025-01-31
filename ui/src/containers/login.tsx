"use client";

import { Box, Flex, Image } from "@raystack/apsara";
import { Header, MagicLink } from "@raystack/frontier/react";
import { useContext } from "react";
import PageTitle from "~/components/page-title";
import { AppContext } from "~/contexts/App";

export default function Login() {
  const { config } = useContext(AppContext);

  return (
    <Flex>
      <PageTitle title="Login" />
      <Box style={{ width: "100%" }}>
        <Flex
          direction="column"
          justify="center"
          align="center"
          style={{
            margin: "auto",
            height: "100vh",
            width: "280px",
          }}
        >
          <Flex
            direction="column"
            style={{ width: "100%", gap: "var(--pd-16)" }}
          >
            <Header
              logo={
                <Image
                  alt="logo"
                  src="logo.svg"
                  width={80}
                  height={80}
                  style={{ borderRadius: "var(--pd-8)" }}
                />
              }
              title={`Login to ${config?.title}`}
            />
            <MagicLink open />
          </Flex>
        </Flex>
      </Box>
    </Flex>
  );
}
