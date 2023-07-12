import React, { useCallback, useState } from "react";
import { useShieldContext } from "../contexts/ShieldContext";
import { hasWindow } from "./helper";

const styles = {
  container: {
    width: "100%",
    display: "flex",
    alignItems: "center",
    gap: "8px",
  },

  input: {
    width: "100%",
    padding: "4px 8px",
    fontSize: "10px",
    lineHeight: "20px",
  },

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

  disabled: { backgroundColor: "rgba(225, 229, 236, 0.4)" },
};

type MagicLinkProps = {
  children?: React.ReactNode;
};
export const MagicLink = ({ children, ...props }: MagicLinkProps) => {
  const { client } = useShieldContext();
  const [loading, setLoading] = useState<boolean>(false);
  const [email, setEmail] = useState<string>("");
  const [otp, setOTP] = useState<string>("");
  const [state, setState] = useState<string>("");

  const magicLinkClickHandler = useCallback(async () => {
    setLoading(true);
    try {
      const {
        data: { state },
      } = await client.getMagicLinkAuthStrategyEndpoint(email);
      setState(state);
    } finally {
      setLoading(false);
    }
  }, [email]);

  const OTPVerifyClickHandler = useCallback(async () => {
    setLoading(true);
    await client.verifyMagicLinkAuthStrategyEndpoint(otp, state);

    const searchParams = new URLSearchParams(
      hasWindow() ? window.location.search : ``
    );
    const redirectURL =
      searchParams.get("redirect_uri") || searchParams.get("redirectURL");

    // @ts-ignore
    window.location = redirectURL ? redirectURL : window.location.origin;
  }, [otp, state]);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(event.target.value);
  };

  const handleOTPChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setOTP(event.target.value);
  };

  return (
    <div style={{ ...styles.container, flexDirection: "column" }}>
      {state ? (
        <input
          key={"otp"}
          style={{ ...styles.input, boxSizing: "border-box" }}
          placeholder="enter OTP"
          onChange={handleOTPChange}
        />
      ) : (
        <input
          key={"email"}
          style={{ ...styles.input, boxSizing: "border-box" }}
          placeholder="name@example.com"
          onChange={handleChange}
        />
      )}

      {loading ? (
        <button
          {...props}
          style={{
            ...styles.button,
          }}
        >
          loading...
        </button>
      ) : (
        <button
          {...props}
          style={{
            ...styles.button,
            ...(!email ? styles.disabled : {}),
          }}
          disabled={!email}
          onClick={state ? OTPVerifyClickHandler : magicLinkClickHandler}
        >
          {state ? "Verify OTP" : "Continue"}
        </button>
      )}
    </div>
  );
};
