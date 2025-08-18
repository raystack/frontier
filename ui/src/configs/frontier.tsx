const getFrontierConfig = () => {
  const frontierEndpoint =
    process.env.NEXT_PUBLIC_FRONTIER_URL || "/frontier-api";
  const frontierConnectEndpoint =
    process.env.NEXT_PUBLIC_FRONTIER_CONNECT_URL || "/frontier-connect";

  const currentHost = window?.location?.origin || "http://localhost:3000";
  return {
    frontierEndpoint: frontierConnectEndpoint,
    endpoint: frontierEndpoint,
    redirectLogin: `${currentHost}/login`,
    redirectSignup: `${currentHost}/signup`,
    redirectMagicLinkVerify: `${currentHost}/magiclink-verify`,
    callbackUrl: `${currentHost}/callback`,
  };
};

export const frontierConfig = getFrontierConfig();
