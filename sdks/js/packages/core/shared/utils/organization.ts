import dayjs from 'dayjs';
import { timestampToDate } from './time';
import type { V1Beta1Organization } from '../../api-client';
import type { Organization } from '@raystack/proton/frontier';

/**
 * Transform ConnectRPC Organization to V1Beta1Organization
 * Converts proto timestamp fields to JSON string format
 */
export function transformConnectOrgToV1Beta1(
  connectOrg: Organization
): V1Beta1Organization {
  if (!connectOrg) return connectOrg as V1Beta1Organization;

  const createdAtString = connectOrg.createdAt
    ? dayjs(timestampToDate(connectOrg.createdAt)).toISOString()
    : undefined;

  const updatedAtString = connectOrg.updatedAt
    ? dayjs(timestampToDate(connectOrg.updatedAt)).toISOString()
    : undefined;

  return {
    id: connectOrg.id,
    name: connectOrg.name,
    title: connectOrg.title,
    metadata: connectOrg.metadata as unknown as object,
    state: connectOrg.state,
    avatar: connectOrg.avatar,
    created_at: createdAtString,
    updated_at: updatedAtString
  };
}
