import { Button, Separator, Text, TextField } from "@raystack/apsara";
import React, { useCallback, useState } from "react";
import { useShield } from "../contexts/ShieldContext";

const styles = {
  container: {
    width: "100%",
    display: "flex",
    alignItems: "center",
    gap: "var(--pd-16)",
  },

  button: {
    width: "100%",
  },
  disabled: { opacity: 1 },
};

type MagicLinkProps = {
  children?: React.ReactNode;
};
export const MagicLink = ({ children, ...props }: MagicLinkProps) => {
  const { client, config } = useShield();
  const [visiable, setVisiable] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const [email, setEmail] = useState<string>("");
  const [state, setState] = useState<string>("");

  const magicLinkClickHandler = useCallback(async () => {
    setLoading(true);
    try {
      const {
        data: { state },
      } = await client.getMagicLinkAuthStrategyEndpoint(email);
      const searchParams = new URLSearchParams({ state, email });
      // @ts-ignore
      window.location = `${
        config.redirectMagicLinkVerify
      }?${searchParams.toString()}`;
    } finally {
      setLoading(false);
    }
  }, [email]);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  };

  if (!visiable)
    return (
      <Button
        variant="secondary"
        size="medium"
        onClick={() => setVisiable(true)}
      >
        <Text>Continue with Email</Text>
      </Button>
    );

  return (
    <div style={{ ...styles.container, flexDirection: "column" }}>
      <Separator />
      <TextField
        size="medium"
        key={"email"}
        placeholder="name@example.com"
        onChange={handleChange}
      />

      <Button
        size="medium"
        variant="primary"
        {...props}
        style={{
          ...styles.button,
          ...(!email ? styles.disabled : {}),
        }}
        disabled={!email}
        onClick={magicLinkClickHandler}
      >
        <Text style={{ color: "var(--foreground-inverted)" }}>
          {loading ? "loading..." : "Continue with Email"}
        </Text>
      </Button>
    </div>
  );
};
