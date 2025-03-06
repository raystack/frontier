"use client";

import { Flex, Image } from "@raystack/apsara/v1";
import { MagicLinkVerify } from "@raystack/frontier/react";

export default function MagicLink() {
  return (
    <Flex
      align="center"
      justify="center"
      style={{ height: "100vh", textAlign: "center" }}
    >
      <MagicLinkVerify
        logo={
          <Image
            alt="logo"
            src="logo.svg"
            width={80}
            height={80}
            style={{ borderRadius: "var(--pd-8)" }}
          />
        }
      />
    </Flex>
  );
}
