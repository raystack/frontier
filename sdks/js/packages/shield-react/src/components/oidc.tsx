import { Button, Text } from "@raystack/apsara";
import React from "react";
const styles = {
  button: {
    width: "100%",
  },
};

type ButtonProps = React.HTMLProps<HTMLButtonElement> & {
  children?: React.ReactNode;
};

export const OIDCButton = ({
  children,
  type = "button",
  onClick,
}: ButtonProps) => (
  <Button
    size="medium"
    variant="secondary"
    style={styles.button}
    onClick={onClick}
  >
    <Text>Continue with {children}</Text>
  </Button>
);
