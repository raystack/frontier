import React from "react";
import { Flex } from "@raystack/apsara";

type SheetFooterProps = {
  children?: React.ReactNode;
};

export function SheetFooter({ children }: SheetFooterProps) {
  // @ts-ignore
  return <Flex style={styles.footer}>{children}</Flex>;
}

const styles = {
  footer: {
    bottom: 0,
    left: 0,
    right: 0,
    position: "absolute",
    justifyContent: "space-between",
    padding: "18px 32px",
    borderTop: "1px solid var(--rs-color-border-base-primary)",
  },
};
