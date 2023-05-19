export type Group = {
  id: string;
  name: string;
  slug: string;
  orgId: string;
  metadata?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
};
