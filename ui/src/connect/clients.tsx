import { createClient } from "@connectrpc/connect";
import { AdminService, FrontierService } from "@raystack/proton/frontier";
import { createFrontierTransport } from "./transport";

interface ClientOptions {
  useBinary?: boolean;
}

export const clients = {
  admin: (options: ClientOptions = {}) => {
    const transport = createFrontierTransport({ useBinary: options.useBinary });
    return createClient(AdminService, transport);
  },
  frontier: (options: ClientOptions = {}) => {
    const transport = createFrontierTransport({ useBinary: options.useBinary });
    return createClient(FrontierService, transport);
  },
};