// Re-export Connect Query hooks for convenience
export {
  useQuery,
  useMutation,
  useInfiniteQuery,
  useTransport
} from '@connectrpc/connect-query';

// Re-export Frontier service queries for convenience
export { FrontierServiceQueries } from '@raystack/proton/frontier';

// Re-export React Query hooks
export { useQueryClient } from '@tanstack/react-query';
