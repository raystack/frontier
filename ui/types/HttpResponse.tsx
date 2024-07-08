import type { RpcStatus } from "@raystack/frontier";

export interface HttpResponse extends Response {
  data: unknown;
  error: RpcStatus;
}
