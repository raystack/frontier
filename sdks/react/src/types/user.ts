export type User = {
  id: string;
  name: string;
  slug: string;
  email: string;
  admin?: boolean;
  metadata?: Record<string, string>;
  createdAt?: string;
  updatedAt?: string;
};
