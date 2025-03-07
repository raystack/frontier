const getFrontierConfig = () => {
  const frontierEndpoint =
    process.env.NEXT_PUBLIC_FRONTIER_URL || "/frontier-api";
  const currentHost = window?.location?.origin || "http://localhost:3000";
  return {
    endpoint: frontierEndpoint,
    redirectLogin: `${currentHost}/login`,
    redirectSignup: `${currentHost}/signup`,
    redirectMagicLinkVerify: `${currentHost}/magiclink-verify`,
    callbackUrl: `${currentHost}/callback`,
  };
};

export const frontierConfig = getFrontierConfig();
