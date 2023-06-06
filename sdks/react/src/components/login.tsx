import React from "react";
import { useStrategyContext } from "../contexts/StrategyContext";

type SignedInProps = {};
export const SignedIn = ({}: SignedInProps) => {
  const { strategies = [] } = useStrategyContext();

  return (
    <Container>
      {strategies.map((s, index) => (
        <Button key={index} onClick={() => (window.location.href = s.endpoint)}>
          Login with {s.name}
        </Button>
      ))}
    </Container>
  );
};

type ContainerProps = { children?: React.ReactNode; css?: React.CSSProperties };
export const Container = ({ children, css }: ContainerProps) => (
  <div
    style={{
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      gap: "8px",
      fontSize: "12px",
      ...css,
    }}
  >
    {children}
  </div>
);

type ButtonProps = React.HTMLProps<HTMLButtonElement> & {
  children?: React.ReactNode;
  css?: React.CSSProperties;
};
export const Button = ({
  children,
  css,
  type = "button",
  ...props
}: ButtonProps) => (
  <button
    {...props}
    style={{
      display: "flex",
      alignItems: "center",
      fontSize: "12px",
      lineHeight: "20px",
      padding: "4px 16px",
      fontWeight: "bold",
      borderRadius: "4px",
      border: "2px solid #D0D0D0",
      color: "#222",
      cursor: "pointer",
      ...css,
    }}
  >
    {children}
  </button>
);
