import { createConnectTransport } from "@connectrpc/connect-web";
import type { Transport } from "@connectrpc/connect";

const frontierConnectEndpoint =
  process.env.NEXT_PUBLIC_FRONTIER_CONNECT_URL || "/frontier-connect";

interface TransportOptions {
  useBinary?: boolean;
  baseUrl?: string;
}

export const createFrontierTransport = (
  options: TransportOptions = {},
): Transport => {
  const { useBinary = false, baseUrl = frontierConnectEndpoint } = options;

  return createConnectTransport({
    baseUrl,
    useBinaryFormat: useBinary,
    interceptors: [],
  });
};

export const jsonTransport = createFrontierTransport();
export const binaryTransport = createFrontierTransport({ useBinary: true });
