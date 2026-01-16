"use client";

import { Box, Flex, Image } from "@raystack/apsara";
import { Header, MagicLink } from "@raystack/frontier/react";
import { useContext } from "react";
import PageTitle from "~/components/page-title";
import { AppContext } from "~/contexts/App";
import { defaultConfig } from "~/utils/constants";
import IAMIcon from "~/assets/icons/iam.svg?react";

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
          <Flex direction="column" gap={5} style={{ width: "100%" }}>
            <Header
              logo={
                config?.logo ? (
                  <Image
                    alt="logo"
                    src={config?.logo}
                    width={80}
                    height={80}
                    style={{ borderRadius: "var(--rs-space-3)" }}
                  />
                ) : (
                  <IAMIcon width={80} height={80} />
                )
              }
              title={`Login to ${config?.title || defaultConfig.title}`}
            />
            <MagicLink open />
          </Flex>
        </Flex>
      </Box>
    </Flex>
  );
}
