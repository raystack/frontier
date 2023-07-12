import React from "react";

const styles = {
  button: {
    width: "100%",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",

    fontSize: "10px",
    lineHeight: "20px",

    padding: "4px 8px",
    fontWeight: "bold",
    borderRadius: "4px",

    borderColor: "transparent",
    backgroundColor: "rgb(225, 229, 236)",

    cursor: "pointer",
    color: "inherit",
  },
};

type ButtonProps = React.HTMLProps<HTMLButtonElement> & {
  children?: React.ReactNode;
};

export const OIDCButton = ({
  children,
  type = "button",
  ...props
}: ButtonProps) => (
  <button
    {...props}
    style={{
      ...styles.button,
    }}
  >
    Sign in with {children}
  </button>
);
