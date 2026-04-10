const getFrontierConfig = () => {
  const connectEndpoint =
    process.env.NEXT_PUBLIC_FRONTIER_CONNECT_URL || "/frontier-connect";

  const currentHost = window?.location?.origin || "http://localhost:3000";
  return {
    connectEndpoint,
    redirectLogin: `${currentHost}/login`,
    redirectSignup: `${currentHost}/signup`,
    redirectMagicLinkVerify: `${currentHost}/magiclink-verify`,
    callbackUrl: `${currentHost}/callback`,
  };
};

export const frontierConfig = getFrontierConfig();
