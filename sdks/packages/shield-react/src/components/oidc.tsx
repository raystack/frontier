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

export const OIDCButton = ({ children, type = "button", ...props }: ButtonProps) => (
  <Button {...props} size="medium" style={styles.button}>
    <Text>Sign in with {children}</Text>
  </Button>
);
