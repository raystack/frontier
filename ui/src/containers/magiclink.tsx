"use client";

import { Flex, Image } from "@raystack/apsara/v1";
import { MagicLinkVerify } from "@raystack/frontier/react";
import IAMIcon from "~/assets/icons/iam.svg?react";
import { AppContext } from "~/contexts/App";
import { useContext } from "react";

export default function MagicLink() {
  const { config } = useContext(AppContext);

  return (
    <Flex
      align="center"
      justify="center"
      style={{ height: "100vh", textAlign: "center" }}
    >
      <MagicLinkVerify
        logo={
          config?.logo ? (
            <Image
              alt="logo"
              src={config?.logo}
              width={80}
              height={80}
              style={{ borderRadius: "var(--pd-8)" }}
            />
          ) : (
            <IAMIcon width={80} height={80} />
          )
        }
      />
    </Flex>
  );
}
