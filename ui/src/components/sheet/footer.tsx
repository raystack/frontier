import { Flex } from "@raystack/apsara";
import { CSS } from "@stitches/react";

type SheetFooterProps = {
  children?: React.ReactNode;
  css?: CSS;
};

export function SheetFooter({ children, css }: SheetFooterProps) {
  return <Flex css={{ ...styles.footer, ...css }}>{children}</Flex>;
}

const styles = {
  footer: {
    bottom: 0,
    left: 0,
    right: 0,
    position: "absolute",
    justifyContent: "space-between",
    padding: "18px 32px",
    borderTop: "1px solid $gray4",
  },
};
